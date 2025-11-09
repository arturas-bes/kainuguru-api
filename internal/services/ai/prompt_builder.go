package ai

import (
	"fmt"
	"strings"
	"time"
)

// PromptBuilder builds high-quality prompts for flyer vision extraction.
// Instruction language is EN. OCR text in outputs must stay Lithuanian exactly as printed.
type PromptBuilder struct {
	storeContext map[string]string
	categories   []string
}

// NewPromptBuilder creates a new prompt builder.
func NewPromptBuilder() *PromptBuilder {
	return &PromptBuilder{
		storeContext: map[string]string{
			"iki":    "IKI (LT grocery). Common visual tags: SUPER KAINA, TIK, MEILĖ IKI (loyalty hearts), IKI EXPRESS, red percentage badges.",
			"maxima": "MAXIMA (LT grocery).",
			"rimi":   "RIMI (LT grocery).",
		},
		categories: []string{
			"mėsa ir žuvis", "pieno produktai", "duona ir konditerija",
			"vaisiai ir daržovės", "gėrimai", "šaldyti produktai",
			"konservai", "kruopos ir makaronai", "saldumynai",
			"higienos prekės", "namų ūkio prekės", "alkoholiniai gėrimai",
		},
	}
}

// Schema returns the JSON schema description we expect from the model.
func (pb *PromptBuilder) Schema() string {
	return `{
  "page_meta": {
    "store_code": "iki|maxima|rimi|...(lowercase exact)",
    "currency": "EUR",
    "locale": "lt-LT",
    "valid_from": "YYYY-MM-DD|null",
    "valid_to": "YYYY-MM-DD|null",
    "page_number": 1,
    "detected_text_sample": "raw OCR snippet near main prices/percents"
  },
  "promotions": [
    {
      "promotion_type": "single_product|category|brand_line|equipment|bundle|loyalty",
      "name_lt": "EXACT Lithuanian text printed near the price/percent (no translation, no paraphrase). If unreadable -> null",
      "brand": "string|null",
      "category_guess_lt": "one from fixed list|null",
      "unit": "kg|g|l|ml|vnt.|pak.|null",
      "unit_size": "e.g., '125 g'|'1 kg'|null",
      "price_eur": "X,XX €|null",
      "original_price_eur": "X,XX €|null",
      "price_per_unit_eur": "X,XX €|null",
      "discount_pct": "integer|null",
      "discount_text": "e.g., '-25 %' as printed|null",
      "discount_type": "percentage|absolute|bundle|loyalty|null",
      "special_tags": ["SUPER KAINA","TIK","MEILĖ IKI","IKI EXPRESS","1+1","2+1","3+1","..."],
      "loyalty_required": "true|false",
      "bundle_details": "e.g., '1+1','2+1','3 už 2'|null",
      "bounding_box": {"x": 0.0, "y": 0.0, "width": 0.0, "height": 0.0},
      "confidence": 0.0
    }
  ],
  "warnings": []
}`
}

// ---- Core prompts -----------------------------------------------------------

// ProductExtractionPrompt now returns the UNIFIED SCHEMA, not legacy "products".
func (pb *PromptBuilder) ProductExtractionPrompt(storeCode string, pageNumber int) string {
	return fmt.Sprintf(
		`ROLE
You extract promotion modules from a Lithuanian grocery flyer image.

OUTPUT
Return ONE JSON object matching the schema below. Strict JSON. No markdown. No commentary.

WHAT TO CAPTURE
A "promotion" is one rectangular module showing any of: a price (€), a percent badge, a bundle (1+1/2+1), or a loyalty tag. Treat group/category callouts (e.g., "Vytintiems mėsos gaminiams -30%%") as a single "category" promotion. If only price exists, "single_product". If only percent exists, set "discount_pct" and keep "price_eur" null. Never invent missing numbers.

ANTI-HALLUCINATION RULES
- Use EXACT Lithuanian text as printed around that price/percent for "name_lt". If the product name is unreadable, set name_lt = null or use the short category headline EXACTLY as printed (e.g., "Makaronams ir užpilamiems makaronams"), not a guessed SKU.
- Do NOT infer flavors/weights/brands unless they are clearly readable near the price/percent inside the same module.
- Do NOT copy a visible price to other modules. Each module must be justified by its own local glyphs (€, percent, bundle).
- If a module shows a loyalty heart/MEILĖ IKI only, set discount_type="loyalty" and loyalty_required=true, with other fields null unless printed.
- If the top of the page is a banner, ignore it unless it contains an actual price/percent/bundle.

NORMALIZATION
- Prices must be "X,XX €". If you read "0 99 €", normalize to "0,99 €". If no €, set price_eur=null.
- Percent must be integer without the sign in "discount_pct", but include the printed form (e.g., "-25 %%") in "discount_text".

STORE: %s | PAGE: %d
SCHEMA
%s`,
		strings.ToLower(storeCode), pageNumber, pb.Schema(),
	)
}

// DetectionPrompt (pass 1) – find modules + coarse fields.
func (pb *PromptBuilder) DetectionPrompt(storeCode string, pageNumber int) string {
	return fmt.Sprintf(`PASS 1: DETECT MODULES

Task: List EVERY rectangular promotion module that shows either a € price, a %% badge, a bundle (1+1/2+1/3+1), or a loyalty tag. For each, return:
- promotion_type (guess)
- name_lt (exact printed headline near price/percent if readable; else null)
- discount_pct OR price_eur (whichever actually printed; never invent both)
- discount_text (as printed if percent)
- loyalty_required (true/false based on hearts/MEILĖ IKI text)
- special_tags
- bounding_box (normalized 0..1)
- confidence

Output strict JSON following the schema. No markdown. Do not drop small corner modules.

STORE: %s | PAGE: %d
SCHEMA
%s`, strings.ToLower(storeCode), pageNumber, pb.Schema())
}

// FillDetailsPrompt (pass 2) – enrich modules provided via bounding boxes.
func (pb *PromptBuilder) FillDetailsPrompt(storeCode string, pageNumber int) string {
	return fmt.Sprintf(`PASS 2: FILL DETAILS

You are given the page image and a JSON list named PROMOTION_BOXES that contains bounding boxes and coarse data found in pass 1.
For each box, read ONLY inside that rectangle and fill or correct fields:
- brand, unit, unit_size
- price_eur, original_price_eur, price_per_unit_eur
- discount_pct and discount_text
- discount_type: percentage|absolute|bundle|loyalty
- category_guess_lt from: [%s]
- special_tags exactly as printed: {"SUPER KAINA","TIK","MEILĖ IKI","IKI EXPRESS","1+1","2+1","3+1",...}

Rules:
- NEVER guess values not printed inside the box.
- If a field is unreadable: null.
- Keep name_lt EXACTLY as printed; if only a category headline is present, use that text verbatim.
- Normalize price to "X,XX €". Normalize percent to integer in discount_pct and preserved form in discount_text.
- Return the SAME number of promotions as PROMOTION_BOXES, in the SAME order, each with a bounding_box.

STORE: %s | PAGE: %d
SCHEMA
%s`,
		strings.Join(pb.categories, ", "),
		strings.ToLower(storeCode), pageNumber, pb.Schema(),
	)
}

// ---- Utilities / secondary prompts -----------------------------------------

// TextExtractionPrompt – English instructions; keep Lithuanian text.
func (pb *PromptBuilder) TextExtractionPrompt(storeCode string) string {
	storeContext := pb.getStoreContext(storeCode)
	return fmt.Sprintf(`Extract ALL legible text from this %s flyer page. Preserve Lithuanian exactly.

CONTEXT:
%s

Return strict JSON:
{
  "header_text": "…",
  "products_text": "…",
  "prices_text": "…",
  "dates_text": "…",
  "promotional_text": "…",
  "other_text": "…"
}

Notes:
- Keep diacritics.
- Do not normalize numbers.
- Include validity date ranges if present.`,
		strings.ToUpper(storeCode), storeContext)
}

// ValidationPromptV2 – strict schema validation/repair (kept for optional use).
func (pb *PromptBuilder) ValidationPromptV2(extractedData string) string {
	prompt := `You will receive JSON that should match the flyer schema. Validate and repair it.

INPUT:
%s

CHECKS
- price_eur / original_price_eur must be "X,XX €" or null. Convert "0 99 €" -> "0,99 €".
- discount_pct is integer 1..99 or null; keep printed form in discount_text if present.
- promotion_type ∈ {single_product, category, brand_line, equipment, bundle, loyalty}.
- Remove obvious non-promotions (legal notes, page legends).
- If both price_eur and original_price_eur exist and original < price, keep both but add a warning.
- Normalize unit/unit_size to [kg,g,l,ml,vnt.,pak.].
- Extract valid_from/valid_to into ISO if present in the text.

OUTPUT
Return the same JSON schema plus a "warnings" array describing fixes. JSON only.`
	return fmt.Sprintf(prompt, extractedData)
}

// Legacy prompt (kept for compatibility if needed elsewhere).
func (pb *PromptBuilder) ValidationPrompt(extractedData string) string {
	prompt := `Patikrink ir pakoreguok ištrauktų produktų duomenis.

GAUTAS DUOMENYS:
%s

TIKRINIMO KRITERIJAI:
1. Kainų formatai turi būti teisingi (pvz., "1,99 €", "2.50 €")
2. Produktų pavadinimai turi būti lietuviškai
3. Mato vienetai turi būti standartiniai (kg, g, l, ml, vnt.)
4. Kategorijos turi atitikti šiuos: %s
5. Nuolaidos informacija turi būti aiški

UŽDUOTIS:
Pataisyk klaidingas kainas, pavadinimus ir kategorijas. Pašalink produktus be kainų.

FORMATAS:
Grąžink pataisytą JSON su papildomu lauku "validation_notes" kiekvienam produktui:
{
  "products": [...],
  "removed_products": [...]
}`
	return fmt.Sprintf(prompt, extractedData, strings.Join(pb.categories, ", "))
}

func (pb *PromptBuilder) CategoryClassificationPrompt(productName string) string {
	return fmt.Sprintf(`Classify the Lithuanian text below into ONE of:
[%s]

TEXT: "%s"

Return: {"category":"...", "confidence":0.00}`,
		strings.Join(pb.categories, ", "), productName)
}

func (pb *PromptBuilder) PriceAnalysisPrompt(products []string) string {
	productsText := strings.Join(products, "\n")
	return fmt.Sprintf(`Input: final promotions JSON
%s

TASK
- Count items with price_eur vs percent-only.
- Average discount_pct over items that have it.
- Compute price_per_unit_eur where price and unit_size exist but PPU is missing.
- List format issues for prices not matching "X,XX €".

OUTPUT
{
  "summary": {"total_promotions":N,"with_price":N,"with_percent_only":N,"avg_discount_pct":number|null},
  "repairs":[{"index":i,"field":"price_per_unit_eur","value":"X,XX €","note":"computed from ..."}],
  "format_issues":["..."]
}`, productsText)
}

func (pb *PromptBuilder) getStoreContext(storeCode string) string {
	if context, ok := pb.storeContext[strings.ToLower(storeCode)]; ok {
		return context
	}
	return "Lietuvos prekybos tinklas"
}

func (pb *PromptBuilder) LayoutAnalysisPrompt() string {
	return `Analizuok šio leidinio puslapio išdėstymą ir struktūrą.

UŽDUOTIS:
Identifikuok ir apibūdink:

1. Puslapio struktūrą (antraštės, skyriai, kolonos)
2. Produktų išdėstymą (eilės, grupės, kategorijos)
3. Kainų pozicionavimą
4. Akcijų ir nuolaidų išryškinimą
5. Prekių ženklų ir kategorijų žymėjimą

FORMATAS:
{
  "layout_analysis": {
    "page_structure": "…",
    "product_layout": "…",
    "sections": [{"type":"…","content":"…","position":"…"}],
    "special_offers": [{"type":"…","description":"…","visual_emphasis":"…"}]
  }
}`
}

func (pb *PromptBuilder) QualityCheckPrompt(extractedData string, _ string) string {
	return fmt.Sprintf(`Patikrink ištrauktų duomenų kokybę palyginti su originaliu vaizdu.

IŠTRAUKTI DUOMENYS:
%s

FORMATAS:
{
  "quality_score": 0.85,
  "completeness": 0.90,
  "accuracy": 0.80,
  "consistency": 0.85,
  "issues_found": [{"type":"…","description":"…","severity":"high|medium|low","suggestion":"…"}],
  "missing_products": skaičius,
  "recommendations": ["…","…"]
}`, extractedData)
}

func (pb *PromptBuilder) BuildCustomPrompt(task, context, requirements string) string {
	ts := time.Now().Format("2006-01-02 15:04")
	return fmt.Sprintf(`UŽDUOTIS: %s

KONTEKSTAS:
%s

REIKALAVIMAI:
%s

KALBOS SPECIFIKA:
- Visas tekstas turi būti lietuvių kalba
- Išlaikyk originalų rašybą ir skyrybą
- Prekių pavadinimus palik kaip originaliai parašyta

LAIKO ŽYMA: %s

Atlikk užduotį tiksliai pagal pateiktus reikalavimus.`, task, context, requirements, ts)
}

func (pb *PromptBuilder) GetAvailableCategories() []string { return pb.categories }
func (pb *PromptBuilder) AddStoreContext(storeCode, context string) {
	pb.storeContext[strings.ToLower(storeCode)] = context
}
func (pb *PromptBuilder) GetSupportedStores() []string {
	stores := make([]string, 0, len(pb.storeContext))
	for s := range pb.storeContext {
		stores = append(stores, s)
	}
	return stores
}
