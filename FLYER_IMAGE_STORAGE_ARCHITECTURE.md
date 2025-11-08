# Flyer Image Storage Architecture
## MVP File System Design (Ultra-Simplified)

**Date:** 2025-11-08  
**Status:** ✅ FINAL ARCHITECTURE - Maximum Simplicity  
**Context:** MVP will use file system storage next to project in public folder

---

## Executive Summary

This document outlines an **ultra-simple folder structure** for storing shop flyer images:
- **Only stores CURRENT + PREVIOUS flyer** (2 flyers max per store)
- Aligns with existing scraper logic
- Uses database URLs instead of complex folder hierarchies
- Automatic cleanup when new flyer arrives
- Stores page 1 as preview for UI
- Keeps folder names lowercase and hyphenated

### Storage Policy
- ✅ **Current Flyer**: Active flyer being displayed
- ✅ **Previous Flyer**: Last week's flyer (one version back)
- ❌ **Archive**: No long-term archive needed
- 🗑️ **Auto-cleanup**: Delete old flyers when new one arrives

---

## Design Philosophy

### Key Insight: Store URLs in Database, Not File Paths
Instead of complex folder structures, we store **full URLs** in the `flyer_pages.image_url` field:
- Scraper downloads images to simple folder structure
- Database stores the final URL for each page
- Active/expired is managed by `flyers.is_archived` flag
- **No need to move files** when archiving

### Benefits
- ✅ **Simple**: Flat folder structure, easy to understand
- ✅ **Flexible**: URLs can change without moving files
- ✅ **Fast**: No complex path calculations
- ✅ **Maintainable**: Easy to debug and manage
- ✅ **Scraper-aligned**: Works with existing `PageInfo.LocalPath`

---

## Proposed Folder Structure

### Directory Hierarchy (Ultra-Simple - Max 2 Flyers Per Store)

```
/Users/arturas/Dev/kainuguru_all/
└── kainuguru-public/                      # Public assets folder
    └── flyers/                            # Root flyer storage
        ├── iki/                           # Store code (lowercase)
        │   ├── 2025-11-10-iki-savaitinis/ # CURRENT: This week's flyer
        │   │   ├── page-1.jpg
        │   │   ├── page-2.jpg
        │   │   └── page-3.jpg
        │   └── 2025-11-03-iki-savaitinis/ # PREVIOUS: Last week's flyer
        │       ├── page-1.jpg
        │       └── page-2.jpg
        │       # 2025-10-27 flyer DELETED when 2025-11-10 arrived
        ├── maxima/
        │   ├── 2025-11-10-maxima-savaitinis/ # CURRENT
        │   └── 2025-11-03-maxima-savaitinis/ # PREVIOUS
        ├── rimi/
        │   ├── 2025-11-10-rimi-savaitinis/   # CURRENT
        │   └── 2025-11-03-rimi-savaitinis/   # PREVIOUS
        ├── lidl/
        │   ├── 2025-11-10-lidl-savaitinis/   # CURRENT
        │   └── 2025-11-03-lidl-savaitinis/   # PREVIOUS
        └── norfa/
            ├── 2025-11-10-norfa-savaitinis/   # CURRENT
            └── 2025-11-03-norfa-savaitinis/   # PREVIOUS
```

### Storage Limits
- **Maximum per store**: 2 folders (current + previous)
- **When new flyer arrives**: Delete oldest folder
- **Total storage**: ~5 stores × 2 flyers × 3MB = **30MB** (tiny!)

### Folder Naming Convention

**Pattern:**
```
{YYYY-MM-DD}-{store}-{title-slug}
```

**Rules:**
1. Date: `YYYY-MM-DD` (valid_from date, ISO 8601)
2. Store: lowercase store code (`iki`, `maxima`, `rimi`, etc.)
3. Title: slugified title (lowercase, hyphens, max 30 chars)
4. No special characters except hyphens
5. Sortable: older flyers naturally sort first

**Examples:**
- `2025-11-03-iki-savaitinis`
- `2025-11-03-maxima-savaitinis`
- `2025-10-20-rimi-specialus-pasiulymas`
- `2024-12-20-lidl-kaledinis`

### Why This Structure?

#### 1. **Simplicity First** ✅
- No nested folders (active/archived)
- No week identifiers
- No flyer IDs in paths
- Just: store → date-based folders

#### 2. **Date-First Sorting** ✅
- Natural chronological order
- Easy to find recent flyers
- Easy to cleanup old flyers (by date prefix)

#### 3. **Two-Flyer Limit** ✅
- Only keep current + previous flyer
- Minimal disk space usage (~30MB total)
- Fast cleanup (just delete oldest folder)
- No long-term storage needed

#### 4. **Human-Readable** ✅
- Clear folder names
- Easy to debug
- Easy to backup/restore
- Easy to migrate

#### 5. **Scraper Integration** ✅
- Aligns with `PageInfo.LocalPath`
- Simple path construction
- Easy to download and save

---

## URL Structure

### Public URL Pattern (Simple)

```
https://yourdomain.com/flyers/{store}/{folder-name}/page-{number}.jpg
```

### Examples

**Current Flyer - Page 1 (Preview):**
```
https://yourdomain.com/flyers/iki/2025-11-03-iki-savaitinis/page-1.jpg
```

**Current Flyer - All Pages:**
```
https://yourdomain.com/flyers/iki/2025-11-03-iki-savaitinis/page-1.jpg
https://yourdomain.com/flyers/iki/2025-11-03-iki-savaitinis/page-2.jpg
https://yourdomain.com/flyers/iki/2025-11-03-iki-savaitinis/page-3.jpg
```

**Old Flyer (Still Accessible, Just Marked Archived in DB):**
```
https://yourdomain.com/flyers/maxima/2024-10-15-maxima-specialus/page-1.jpg
```

### URL Benefits
- ✅ **Simple**: No complex path segments
- ✅ **Stable**: URLs don't change when archiving
- ✅ **Readable**: Clear meaning from URL
- ✅ **Cacheable**: Static file URLs
- ✅ **SEO-friendly**: Date and store in path

---

## Database Integration

### Key Principle: URLs Stored in Database

**The database stores complete URLs, not relative paths:**
- `flyers.source_url`: Original PDF/flyer URL from store website
- `flyer_pages.image_url`: **Full public URL to our stored image**

**Example Database Values:**
```sql
-- Flyer record
INSERT INTO flyers (store_id, title, valid_from, valid_to, source_url, is_archived)
VALUES (1, 'IKI savaitinis', '2025-11-03', '2025-11-09', 
        'https://iki.lt/wp-content/uploads/2025/11/03/iki-leidinys.pdf', false);

-- Flyer page records (with full URLs)
INSERT INTO flyer_pages (flyer_id, page_number, image_url)
VALUES 
  (123, 1, 'https://yourdomain.com/flyers/iki/2025-11-03-iki-savaitinis/page-1.jpg'),
  (123, 2, 'https://yourdomain.com/flyers/iki/2025-11-03-iki-savaitinis/page-2.jpg'),
  (123, 3, 'https://yourdomain.com/flyers/iki/2025-11-03-iki-savaitinis/page-3.jpg');
```

### Flyer Model Enhancement

```go
// internal/models/flyer.go

// GetFolderName returns the folder name for this flyer's images
func (f *Flyer) GetFolderName() string {
    // Format: YYYY-MM-DD-store-title-slug
    dateStr := f.ValidFrom.Format("2006-01-02")
    storeCode := strings.ToLower(f.Store.Code)
    titleSlug := slugify(f.Title)
    
    return fmt.Sprintf("%s-%s-%s", dateStr, storeCode, titleSlug)
}

// GetImageBasePath returns the base directory path for this flyer's images
func (f *Flyer) GetImageBasePath() string {
    storeCode := strings.ToLower(f.Store.Code)
    folderName := f.GetFolderName()
    
    return fmt.Sprintf("flyers/%s/%s", storeCode, folderName)
}

// GetImageURL returns the full URL for a specific page
func (f *Flyer) GetImageURL(pageNumber int, baseURL string) string {
    basePath := f.GetImageBasePath()
    return fmt.Sprintf("%s/%s/page-%d.jpg", baseURL, basePath, pageNumber)
}

// GetPreviewImageURL returns the URL for page 1 (preview/thumbnail)
func (f *Flyer) GetPreviewImageURL(baseURL string) string {
    return f.GetImageURL(1, baseURL)
}

// slugify converts title to URL-friendly slug
func slugify(title string) string {
    // Convert to lowercase
    slug := strings.ToLower(title)
    
    // Replace Lithuanian characters
    replacements := map[string]string{
        "ą": "a", "č": "c", "ę": "e", "ė": "e",
        "į": "i", "š": "s", "ų": "u", "ū": "u", "ž": "z",
    }
    for lt, en := range replacements {
        slug = strings.ReplaceAll(slug, lt, en)
    }
    
    // Remove non-alphanumeric characters (keep spaces for now)
    reg := regexp.MustCompile("[^a-z0-9 ]")
    slug = reg.ReplaceAllString(slug, "")
    
    // Replace spaces with hyphens
    slug = strings.ReplaceAll(slug, " ", "-")
    
    // Remove multiple consecutive hyphens
    reg = regexp.MustCompile("-+")
    slug = reg.ReplaceAllString(slug, "-")
    
    // Trim hyphens from start/end
    slug = strings.Trim(slug, "-")
    
    // Limit length
    if len(slug) > 30 {
        slug = slug[:30]
        // Trim trailing hyphen if cut in middle of word
        slug = strings.TrimRight(slug, "-")
    }
    
    return slug
}
```

### FlyerPage Model Enhancement

```go
// internal/models/flyer_page.go

// GetLocalPath returns the local file path for this page (for saving)
func (fp *FlyerPage) GetLocalPath(baseStoragePath string) string {
    if fp.Flyer == nil {
        return ""
    }
    basePath := fp.Flyer.GetImageBasePath()
    fileName := fmt.Sprintf("page-%d.jpg", fp.PageNumber)
    
    return filepath.Join(baseStoragePath, basePath, fileName)
}

// GetPublicURL returns the public URL (reads from database field)
func (fp *FlyerPage) GetPublicURL() string {
    // ImageURL is already stored as full URL in database
    return *fp.ImageURL
}

// IsFirstPage checks if this is page 1 (used for preview)
func (fp *FlyerPage) IsFirstPage() bool {
    return fp.PageNumber == 1
}
```

---

## Service Layer Implementation

### File Storage Service (Simplified)

Create a new service to handle file operations:

```go
// internal/services/storage/flyer_storage.go
package storage

import (
    "context"
    "fmt"
    "io"
    "os"
    "path/filepath"
    
    "github.com/kainuguru/kainuguru-api/internal/models"
)

type FlyerStorageService interface {
    SaveFlyerPage(ctx context.Context, flyer *models.Flyer, pageNumber int, imageData io.Reader) (string, error)
    GetFlyerPageURL(flyer *models.Flyer, pageNumber int) string
    DeleteFlyer(ctx context.Context, flyer *models.Flyer) error
    EnforceStorageLimit(ctx context.Context, storeCode string) error // Keep only 2 newest flyers
}

type fileSystemStorage struct {
    basePath  string // e.g., "/path/to/kainuguru-public"
    publicURL string // e.g., "https://yourdomain.com"
}

func NewFileSystemStorage(basePath, publicURL string) FlyerStorageService {
    return &fileSystemStorage{
        basePath:  basePath,
        publicURL: publicURL,
    }
}

func (s *fileSystemStorage) SaveFlyerPage(ctx context.Context, flyer *models.Flyer, pageNumber int, imageData io.Reader) (string, error) {
    // Construct file path
    relativePath := flyer.GetImageBasePath()
    fullPath := filepath.Join(s.basePath, relativePath)
    
    // Create directories if they don't exist
    if err := os.MkdirAll(fullPath, 0755); err != nil {
        return "", fmt.Errorf("failed to create directory: %w", err)
    }
    
    // Create file
    fileName := fmt.Sprintf("page-%d.jpg", pageNumber)
    filePath := filepath.Join(fullPath, fileName)
    
    file, err := os.Create(filePath)
    if err != nil {
        return "", fmt.Errorf("failed to create file: %w", err)
    }
    defer file.Close()
    
    // Copy image data
    _, err = io.Copy(file, imageData)
    if err != nil {
        return "", fmt.Errorf("failed to write image data: %w", err)
    }
    
    // Return the public URL
    publicURL := s.GetFlyerPageURL(flyer, pageNumber)
    return publicURL, nil
}

func (s *fileSystemStorage) GetFlyerPageURL(flyer *models.Flyer, pageNumber int) string {
    relativePath := flyer.GetImageBasePath()
    fileName := fmt.Sprintf("page-%d.jpg", pageNumber)
    return fmt.Sprintf("%s/%s/%s", s.publicURL, relativePath, fileName)
}

func (s *fileSystemStorage) DeleteFlyer(ctx context.Context, flyer *models.Flyer) error {
    flyerPath := filepath.Join(s.basePath, flyer.GetImageBasePath())
    return os.RemoveAll(flyerPath)
}

func (s *fileSystemStorage) EnforceStorageLimit(ctx context.Context, storeCode string) error {
    storePath := filepath.Join(s.basePath, "flyers", strings.ToLower(storeCode))
    
    // Read all flyer folders for this store
    flyerFolders, err := os.ReadDir(storePath)
    if err != nil {
        return fmt.Errorf("failed to read store directory: %w", err)
    }
    
    // Filter only directories and sort by date (folder name starts with date)
    var folders []string
    for _, folder := range flyerFolders {
        if folder.IsDir() {
            folders = append(folders, folder.Name())
        }
    }
    
    // Sort folders (date prefix ensures chronological order)
    sort.Strings(folders)
    
    // Keep only the 2 newest folders
    const maxFolders = 2
    if len(folders) > maxFolders {
        // Delete oldest folders
        for i := 0; i < len(folders)-maxFolders; i++ {
            oldFolderPath := filepath.Join(storePath, folders[i])
            if err := os.RemoveAll(oldFolderPath); err != nil {
                log.Error().Err(err).Str("path", oldFolderPath).Msg("Failed to delete old flyer")
            } else {
                log.Info().Str("path", oldFolderPath).Msg("Deleted old flyer (keeping only 2 newest)")
            }
        }
    }
    
    return nil
}
```

---

## Scraper Integration Workflow

### Complete Flow: Scrape → Download → Store → Database

```go
// Simplified scraper workflow with storage service

func processFlyerFromScraper(ctx context.Context, flyerInfo scraper.FlyerInfo, storageService FlyerStorageService, db *bun.DB) error {
    // 1. Create flyer record in database
    flyer := &models.Flyer{
        StoreID:   flyerInfo.StoreID,
        Title:     flyerInfo.Title,
        ValidFrom: flyerInfo.ValidFrom,
        ValidTo:   flyerInfo.ValidTo,
        SourceURL: &flyerInfo.FlyerURL,
        Status:    "pending",
    }
    
    // Load store relation (needed for folder path)
    if err := db.NewSelect().Model(flyer).Relation("Store").WherePK().Scan(ctx); err != nil {
        return err
    }
    
    _, err := db.NewInsert().Model(flyer).Exec(ctx)
    if err != nil {
        return fmt.Errorf("failed to create flyer: %w", err)
    }
    
    // 2. Download PDF and convert to images (or download images directly)
    pages, err := downloadAndProcessFlyer(ctx, flyerInfo.FlyerURL)
    if err != nil {
        return fmt.Errorf("failed to download flyer: %w", err)
    }
    
    // 3. Save each page and store URL in database
    for i, pageData := range pages {
        pageNumber := i + 1
        
        // Save image file and get public URL
        publicURL, err := storageService.SaveFlyerPage(ctx, flyer, pageNumber, pageData)
        if err != nil {
            return fmt.Errorf("failed to save page %d: %w", pageNumber, err)
        }
        
        // Create flyer_page record with URL
        flyerPage := &models.FlyerPage{
            FlyerID:    flyer.ID,
            PageNumber: pageNumber,
            ImageURL:   &publicURL, // Store full URL!
            ExtractionStatus: "pending",
        }
        
        _, err = db.NewInsert().Model(flyerPage).Exec(ctx)
        if err != nil {
            return fmt.Errorf("failed to create flyer page: %w", err)
        }
    }
    
    // 4. Update flyer status
    flyer.Status = "completed"
    flyer.PageCount = &len(pages)
    _, err = db.NewUpdate().Model(flyer).WherePK().Exec(ctx)
    
    return err
}
```

### Scraper PageInfo → Storage

The scraper's `PageInfo.LocalPath` is replaced by our storage service:

**Before (Scraper manages files):**
```go
type PageInfo struct {
    ImageURL   string // External URL from store
    LocalPath  string // Where scraper saved it
}
```

**After (Storage service manages files):**
```go
// Scraper returns image data
imageData := downloadImage(pageInfo.ImageURL)

// Storage service saves and returns public URL
publicURL, err := storageService.SaveFlyerPage(ctx, flyer, pageNumber, imageData)

// Database stores the public URL
flyerPage.ImageURL = &publicURL
```

---

## GraphQL Integration

### Schema Enhancement

```graphql
# internal/graphql/schema/flyer.graphql

type Flyer {
  id: Int!
  storeId: Int!
  store: Store!
  title: String
  validFrom: String!
  validTo: String!
  pageCount: Int
  
  # Image URLs
  previewImageUrl: String!       # Always page 1
  imageBaseUrl: String!           # Base URL for all pages
  pages: [FlyerPage!]!
  
  isArchived: Boolean!
  status: String!
  daysRemaining: Int!
}

type FlyerPage {
  id: Int!
  flyerId: Int!
  pageNumber: Int!
  imageUrl: String!               # Full URL to image
  isFirstPage: Boolean!           # Useful for preview logic
  extractionStatus: String!
}

extend type Query {
  # Get current active flyers for a store
  currentFlyers(storeCode: String!): [Flyer!]!
  
  # Get a specific flyer with all pages
  flyer(id: Int!): Flyer
  
  # Get flyer by store and week
  flyerByWeek(storeCode: String!, week: String!): [Flyer!]!
}
```

### Resolver Implementation

```go
// internal/graphql/resolvers/flyer.go

func (r *queryResolver) CurrentFlyers(ctx context.Context, storeCode string) ([]*model.Flyer, error) {
    baseURL := r.config.PublicBaseURL // e.g., "https://yourdomain.com"
    
    flyers, err := r.services.Flyer.GetCurrentFlyers(ctx, storeCode)
    if err != nil {
        return nil, err
    }
    
    // Enhance with URLs
    for _, flyer := range flyers {
        // URL generation happens in model methods
    }
    
    return flyers, nil
}

// Field resolver for Flyer.previewImageUrl
func (r *flyerResolver) PreviewImageURL(ctx context.Context, obj *model.Flyer) (string, error) {
    baseURL := r.config.PublicBaseURL
    return obj.GetPreviewImageURL(baseURL), nil
}

// Field resolver for Flyer.pages with image URLs
func (r *flyerResolver) Pages(ctx context.Context, obj *model.Flyer) ([]*model.FlyerPage, error) {
    pages, err := r.services.FlyerPage.GetByFlyerID(ctx, obj.ID)
    if err != nil {
        return nil, err
    }
    
    baseURL := r.config.PublicBaseURL
    for _, page := range pages {
        page.ImageURL = page.GetImageURL(baseURL)
    }
    
    return pages, nil
}
```

---

## File Naming Conventions

### Standards

1. **Store Codes**: Lowercase, URL-friendly
   - `iki`, `maxima`, `rimi`, `lidl`, `norfa`
   
2. **Week Format**: ISO 8601 (YYYY-Www)
   - `2025-W45`, `2025-W46`
   
3. **Year/Week in Archives**: Separate for better organization
   - `2025/W45`, `2024/W52`
   
4. **Flyer Folders**: Prefixed with "flyer-" + ID
   - `flyer-123`, `flyer-456`
   
5. **Page Files**: Hyphenated, numbered
   - `page-1.jpg`, `page-2.jpg`, `page-10.jpg`

### Image Format Standards

**MVP Phase (File System):**
- **Primary Format**: JPEG (`.jpg`)
- **Quality**: 85% compression (good balance)
- **Max Width**: 1200px (mobile-friendly)
- **Thumbnails**: Generate 400px wide for list views

**Future Enhancements:**
- WebP format for better compression
- Multiple sizes (thumbnail, medium, full)
- Lazy loading support

---

## Configuration

### Environment Variables

```env
# .env
FLYER_STORAGE_BASE_PATH=/Users/arturas/Dev/kainuguru_all/kainuguru-public
FLYER_STORAGE_TYPE=filesystem  # filesystem | s3 (future)
PUBLIC_BASE_URL=http://localhost:8080
```

### Config Structure

```go
// internal/config/config.go

type Config struct {
    // ... existing fields
    
    Storage StorageConfig `json:"storage"`
}

type StorageConfig struct {
    Type            string `json:"type" env:"FLYER_STORAGE_TYPE" envDefault:"filesystem"`
    BasePath        string `json:"base_path" env:"FLYER_STORAGE_BASE_PATH"`
    PublicBaseURL   string `json:"public_base_url" env:"PUBLIC_BASE_URL"`
    MaxFileSize     int64  `json:"max_file_size" envDefault:"10485760"` // 10MB
    AllowedFormats  []string `json:"allowed_formats" envDefault:"jpg,jpeg,png"`
}
```

---

## HTTP Static File Serving

### Fiber Configuration

```go
// cmd/api/server/server.go

func setupRoutes(app *fiber.App, config *config.Config) {
    // Serve static flyer images
    app.Static("/flyers", config.Storage.BasePath+"/flyers", fiber.Static{
        Compress:      true,
        ByteRange:     true,
        Browse:        false,  // Don't allow directory listing
        Index:         "",
        CacheDuration: 24 * time.Hour,  // Cache for 24 hours
        MaxAge:        86400,           // Browser cache: 1 day
    })
    
    // Rest of routes...
}
```

### Nginx Configuration (Production)

```nginx
# nginx/nginx.conf

server {
    listen 80;
    server_name yourdomain.com;
    
    # Flyer images
    location /flyers/ {
        alias /path/to/kainuguru-public/flyers/;
        
        # Cache active flyers for 1 day
        location ~ /flyers/[^/]+/active/ {
            expires 1d;
            add_header Cache-Control "public, immutable";
        }
        
        # Cache archived flyers for 1 year
        location ~ /flyers/[^/]+/archived/ {
            expires 1y;
            add_header Cache-Control "public, immutable";
        }
        
        # Security
        add_header X-Content-Type-Options "nosniff";
        
        # Performance
        gzip on;
        gzip_types image/jpeg image/png;
    }
    
    # API
    location / {
        proxy_pass http://api:8080;
    }
}
```

---

## Migration Path to Cloud Storage

### Design for Future S3 Migration

The architecture is **cloud-ready** with minimal changes:

```go
// interface stays the same
type FlyerStorageService interface {
    SaveFlyerPage(ctx context.Context, flyer *models.Flyer, pageNumber int, imageData io.Reader) error
    // ... other methods
}

// New S3 implementation
type s3Storage struct {
    client    *s3.Client
    bucket    string
    cdnURL    string
}

func NewS3Storage(client *s3.Client, bucket, cdnURL string) FlyerStorageService {
    return &s3Storage{
        client: client,
        bucket: bucket,
        cdnURL: cdnURL,
    }
}

// Same interface, different implementation
func (s *s3Storage) SaveFlyerPage(ctx context.Context, flyer *models.Flyer, pageNumber int, imageData io.Reader) error {
    key := flyer.GetImageBasePath() + fmt.Sprintf("/page-%d.jpg", pageNumber)
    
    _, err := s.client.PutObject(ctx, &s3.PutObjectInput{
        Bucket: aws.String(s.bucket),
        Key:    aws.String(key),
        Body:   imageData,
    })
    
    return err
}
```

**Key Benefits:**
- ✅ Same folder structure translates to S3 keys
- ✅ Same URL patterns (just different domain)
- ✅ No database changes needed
- ✅ Swap implementation via config

---

## Operational Procedures

### 1. Automatic Cleanup When New Flyer Arrives

```go
// Called automatically during flyer processing
func saveNewFlyer(ctx context.Context, flyerInfo scraper.FlyerInfo) error {
    // 1. Save the new flyer
    flyer := &models.Flyer{
        StoreID:   flyerInfo.StoreID,
        Title:     flyerInfo.Title,
        ValidFrom: flyerInfo.ValidFrom,
        ValidTo:   flyerInfo.ValidTo,
        // ... other fields
    }
    
    _, err := db.NewInsert().Model(flyer).Exec(ctx)
    if err != nil {
        return err
    }
    
    // 2. Save flyer images
    for i, pageData := range pages {
        publicURL, err := storageService.SaveFlyerPage(ctx, flyer, i+1, pageData)
        // ... save to database
    }
    
    // 3. AUTOMATIC CLEANUP: Enforce 2-flyer limit for this store
    if err := storageService.EnforceStorageLimit(ctx, flyer.Store.Code); err != nil {
        log.Error().Err(err).Msg("Failed to enforce storage limit")
        // Don't fail the whole operation
    }
    
    // 4. Mark old flyers as archived in database (but don't delete)
    result, err := db.NewUpdate().
        Model(&models.Flyer{}).
        Set("is_archived = ?", true).
        Set("archived_at = ?", time.Now()).
        Where("store_id = ?", flyer.StoreID).
        Where("id != ?", flyer.ID).  // Not the new one
        Where("is_archived = ?", false).
        Exec(ctx)
    
    return err
}
```

**Key Points:**
- ✅ Cleanup happens automatically when new flyer arrives
- ✅ Always keeps exactly 2 folders per store (current + previous)
- ✅ No manual intervention needed
- ✅ No cron jobs needed for cleanup

### 2. Manual Cleanup (If Needed)

```go
// Run manually if needed (e.g., if automatic cleanup failed)
func cleanupAllStores(ctx context.Context) error {
    stores := []string{"iki", "maxima", "rimi", "lidl", "norfa"}
    
    for _, storeCode := range stores {
        if err := storageService.EnforceStorageLimit(ctx, storeCode); err != nil {
            log.Error().Err(err).Str("store", storeCode).Msg("Failed to cleanup store")
        }
    }
    
    return nil
}
```

### 3. Health Check

```go
func checkStorageHealth(ctx context.Context) error {
    // Check disk space
    var stat syscall.Statfs_t
    if err := syscall.Statfs(storagePath, &stat); err != nil {
        return fmt.Errorf("failed to get disk stats: %w", err)
    }
    
    availableGB := float64(stat.Bavail*uint64(stat.Bsize)) / 1024 / 1024 / 1024
    if availableGB < 5.0 {
        return fmt.Errorf("low disk space: %.2f GB available", availableGB)
    }
    
    // Check that active flyers have page 1 images
    var activeFlyers []*models.Flyer
    err := db.NewSelect().
        Model(&activeFlyers).
        Where("is_archived = ?", false).
        Relation("Store").
        Scan(ctx)
    
    if err != nil {
        return err
    }
    
    missingCount := 0
    for _, flyer := range activeFlyers {
        page1Path := filepath.Join(storagePath, flyer.GetImageBasePath(), "page-1.jpg")
        if _, err := os.Stat(page1Path); os.IsNotExist(err) {
            log.Warn().Int("flyerId", flyer.ID).Msg("Missing page 1 image")
            missingCount++
        }
    }
    
    if missingCount > 0 {
        return fmt.Errorf("%d active flyers missing page 1 images", missingCount)
    }
    
    return nil
}
```

---

## Performance Considerations

### Disk Space Estimation (Ultra-Minimal)

**Assumptions:**
- 5 stores
- 2 flyers per store (current + previous)
- Average 15 pages per flyer
- Average 200KB per page (compressed JPEG)

**Calculation:**
```
Per Flyer:     15 pages × 200KB = 3MB
Per Store:     2 flyers × 3MB = 6MB
Total Storage: 5 stores × 6MB = 30MB
```

**Total Storage Needed:** ~**30MB** (incredibly small!)

**Comparison:**
- ❌ Old approach (2 years): 1.56GB
- ✅ New approach (2 flyers): 30MB
- 💾 **Storage reduction: 98%!**

### Performance Metrics

**File Access:**
- Local filesystem: ~1-5ms
- Served via Fiber: ~10-20ms
- With nginx: ~5-10ms

**Caching Strategy:**
- Browser cache: 1 day (active), 1 year (archived)
- CDN cache (future): Same as browser
- Application cache: Not needed (static files)

---

## Error Handling

### Common Scenarios

1. **Missing Image File**
   ```go
   func getFlyerPage(flyerID, pageNumber int) (string, error) {
       path := getImagePath(flyerID, pageNumber)
       if _, err := os.Stat(path); os.IsNotExist(err) {
           return "", fmt.Errorf("image not found: %w", err)
       }
       return path, nil
   }
   ```

2. **Corrupted Image**
   - Return placeholder image
   - Log for manual review
   - Mark page for reprocessing

3. **Disk Full**
   - Monitor disk space
   - Alert before reaching 80%
   - Cleanup old archives automatically

---

## Testing Strategy

### Unit Tests

```go
func TestFlyerImagePath(t *testing.T) {
    flyer := &models.Flyer{
        ID:        123,
        StoreID:   1,
        ValidFrom: time.Date(2025, 11, 4, 0, 0, 0, 0, time.UTC),
        Store: &models.Store{
            Code: "IKI",
        },
    }
    
    expected := "flyers/iki/active/2025-W45/flyer-123"
    actual := flyer.GetImageBasePath()
    
    assert.Equal(t, expected, actual)
}
```

### Integration Tests

```go
func TestSaveFlyerPage(t *testing.T) {
    tempDir := t.TempDir()
    storage := NewFileSystemStorage(tempDir)
    
    imageData := bytes.NewReader([]byte("fake image data"))
    err := storage.SaveFlyerPage(ctx, flyer, 1, imageData)
    
    assert.NoError(t, err)
    assert.FileExists(t, filepath.Join(tempDir, flyer.GetImageBasePath(), "page-1.jpg"))
}
```

---

## Security Considerations

### Access Control

1. **Public Read**: Anyone can view flyer images
2. **Write Permissions**: Only backend service
3. **Delete Permissions**: Only admin users
4. **Directory Listing**: Disabled

### File Upload Security

```go
func validateImage(imageData io.Reader) error {
    // Check file signature (magic bytes)
    // Validate image format
    // Check file size limits
    // Scan for malware (if needed)
    return nil
}
```

---

## Monitoring & Alerts

### Metrics to Track

1. **Storage Metrics:**
   - Disk usage by store
   - Total file count
   - Average file size

2. **Performance Metrics:**
   - Image serve latency
   - 404 error rate
   - Cache hit ratio

3. **Business Metrics:**
   - Active flyers count
   - Archived flyers count
   - Missing images count

### Alert Rules

- Disk usage > 80%
- Missing image rate > 5%
- Image serve latency > 100ms (p95)

---

## Summary

### ✅ Proposed Structure Benefits

1. **Ultra-Simple**: Flat structure, no complex nesting
2. **Minimal Storage**: Only 30MB total (2 flyers per store)
3. **Auto-Cleanup**: Automatic when new flyer arrives
4. **Fast**: No archive management needed
5. **Performant**: Small footprint, quick access
6. **Maintainable**: Self-managing storage
7. **User-Friendly**: Current + previous flyer always available

### Quick Reference: Path Templates

**All Flyers (Same Pattern):**
```
/flyers/{store}/{YYYY-MM-DD-store-title-slug}/page-{n}.jpg
```

**Examples:**
```
/flyers/iki/2025-11-03-iki-savaitinis/page-1.jpg
/flyers/maxima/2025-11-03-maxima-savaitinis/page-1.jpg
/flyers/rimi/2024-10-15-rimi-specialus/page-1.jpg
```

**Preview (Page 1):**
```
/flyers/{store}/{folder-name}/page-1.jpg
```

**Important:** Active vs Archived is determined by database flag, NOT folder structure!

---

## Next Steps

1. ✅ **Architecture approved** ← YOU ARE HERE
2. ⏳ Create `kainuguru-public/flyers/` directory
3. ⏳ Implement `FlyerStorageService` with `EnforceStorageLimit()`
4. ⏳ Add model helper methods (`GetFolderName()`, `GetImageBasePath()`)
5. ⏳ Update scraper to use storage service
6. ⏳ Configure static file serving (Fiber/Nginx)
7. ⏳ Write tests
8. ⏳ Update documentation

---

## Summary

**Storage Policy:** Current + Previous only (2 flyers per store max)  
**Total Storage:** ~30MB (98% reduction vs archive approach)  
**Cleanup:** Automatic when new flyer arrives  
**Folder Pattern:** `{YYYY-MM-DD}-{store}-{title-slug}`  
**Example:** `/flyers/iki/2025-11-10-iki-savaitinis/page-1.jpg`

---

*Architecture designed for Kainuguru MVP*  
*Date: 2025-11-08*  
*Status: Final - Ultra-Simplified ✅*
