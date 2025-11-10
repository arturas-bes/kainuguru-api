package enrichment

import (
"context"
"fmt"
"time"

"github.com/kainuguru/kainuguru-api/internal/config"
"github.com/kainuguru/kainuguru-api/internal/database"
"github.com/kainuguru/kainuguru-api/internal/models"
"github.com/kainuguru/kainuguru-api/internal/services"
"github.com/kainuguru/kainuguru-api/internal/services/ai"
"github.com/rs/zerolog/log"
)

// ProcessOptions defines options for flyer processing
type ProcessOptions struct {
StoreCode      string
Date           time.Time
ForceReprocess bool
MaxPages       int
BatchSize      int
DryRun         bool
}

// Orchestrator orchestrates the flyer enrichment process
type Orchestrator struct {
db              *database.BunDB
cfg             *config.Config
enrichmentSvc   services.EnrichmentService
}

// NewOrchestrator creates a new enrichment orchestrator instance
func NewOrchestrator(ctx context.Context, db *database.BunDB, cfg *config.Config) (*Orchestrator, error) {
	// Create AI extractor with config from environment
	extractorConfig := ai.ExtractorConfig{
		OpenAIAPIKey:  cfg.OpenAI.APIKey,
		Model:         cfg.OpenAI.Model,
		MaxTokens:     cfg.OpenAI.MaxTokens,
		Temperature:   cfg.OpenAI.Temperature,
		MaxRetries:    cfg.OpenAI.MaxRetries,
		RetryDelay:    2 * time.Second,
		Timeout:       cfg.OpenAI.Timeout,
		EnableCaching: true,
		CacheExpiry:   24 * time.Hour,
		BatchSize:     5,
	}
	aiExtractor := ai.NewProductExtractor(extractorConfig)

// Create service factory
serviceFactory := services.NewServiceFactory(db.DB)

// Create enrichment service
enrichmentSvc := NewService(
db.DB,
serviceFactory.FlyerService(),
serviceFactory.FlyerPageService(),
serviceFactory.ProductService(),
serviceFactory.ProductMasterService(),
aiExtractor,
)

return &Orchestrator{
db:            db,
cfg:           cfg,
enrichmentSvc: enrichmentSvc,
}, nil
}

// ProcessFlyers processes flyers based on provided options
func (o *Orchestrator) ProcessFlyers(ctx context.Context, opts ProcessOptions) error {
log.Info().Msg("Starting flyer processing")

// Get eligible flyers
flyers, err := o.enrichmentSvc.GetEligibleFlyers(ctx, opts.Date, opts.StoreCode)
if err != nil {
return fmt.Errorf("failed to get eligible flyers: %w", err)
}

if len(flyers) == 0 {
log.Info().Msg("No eligible flyers found for processing")
return nil
}

log.Info().Int("count", len(flyers)).Msg("Found eligible flyers")

if opts.DryRun {
return o.dryRun(flyers)
}

return o.processAllFlyers(ctx, flyers, opts)
}

func (o *Orchestrator) dryRun(flyers []*models.Flyer) error {
log.Info().Msg("Dry run mode - listing flyers that would be processed:")

for _, flyer := range flyers {
storeName := "Unknown"
if flyer.Store != nil {
storeName = flyer.Store.Name
}

log.Info().
Int("id", flyer.ID).
Str("store", storeName).
Time("valid_from", flyer.ValidFrom).
Time("valid_to", flyer.ValidTo).
Msg("Would process flyer")
}

return nil
}

func (o *Orchestrator) processAllFlyers(ctx context.Context, flyers []*models.Flyer, opts ProcessOptions) error {
totalProcessed := 0
totalProducts := 0
flyersProcessedCount := 0

for _, flyer := range flyers {
select {
case <-ctx.Done():
log.Info().Msg("Context cancelled, stopping processing")
return ctx.Err()
default:
}

// Check if we've reached the max pages limit before processing next flyer
if opts.MaxPages > 0 && totalProcessed >= opts.MaxPages {
log.Info().
Int("max_pages", opts.MaxPages).
Int("pages_processed", totalProcessed).
Msg("Reached maximum pages limit, stopping")
break
}

storeName := "Unknown"
if flyer.Store != nil {
storeName = flyer.Store.Name
}

log.Info().
Int("flyer_id", flyer.ID).
Str("store", storeName).
Msg("Processing flyer")

// Calculate remaining pages if maxPages is set
remainingPages := 0
if opts.MaxPages > 0 {
remainingPages = opts.MaxPages - totalProcessed
if remainingPages <= 0 {
log.Info().Msg("No remaining pages to process")
break
}
}

stats, err := o.enrichmentSvc.ProcessFlyer(ctx, flyer, services.EnrichmentOptions{
ForceReprocess: opts.ForceReprocess,
MaxPages:       remainingPages,
BatchSize:      opts.BatchSize,
})

if err != nil {
log.Error().
Err(err).
Int("flyer_id", flyer.ID).
Msg("Failed to process flyer")
continue
}

log.Info().
Int("flyer_id", flyer.ID).
Int("pages_processed", stats.PagesProcessed).
Int("products_extracted", stats.ProductsExtracted).
Int("pages_failed", stats.PagesFailed).
Float64("avg_confidence", stats.AvgConfidence).
Dur("duration", stats.Duration).
Msg("Flyer processing completed")

totalProcessed += stats.PagesProcessed
totalProducts += stats.ProductsExtracted
flyersProcessedCount++

// Stop if we've reached the max pages limit after processing this flyer
if opts.MaxPages > 0 && totalProcessed >= opts.MaxPages {
log.Info().
Int("max_pages", opts.MaxPages).
Int("pages_processed", totalProcessed).
Msg("Reached maximum pages limit after flyer processing")
break
}
}

log.Info().
Int("flyers_processed", flyersProcessedCount).
Int("pages_processed", totalProcessed).
Int("products_extracted", totalProducts).
Msg("Processing summary")

return nil
}
