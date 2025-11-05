package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

// ProductCategory represents a predefined product category
type ProductCategory struct {
	bun.BaseModel `bun:"table:product_categories,alias:pc"`

	ID             int    `bun:"id,pk,autoincrement" json:"id"`
	Name           string `bun:"name,unique,notnull" json:"name"`
	NormalizedName string `bun:"normalized_name,notnull" json:"normalized_name"`
	ParentID       *int   `bun:"parent_id" json:"parent_id"`

	// Display information
	DisplayNameLT string  `bun:"display_name_lt,notnull" json:"display_name_lt"`
	DisplayNameEN *string `bun:"display_name_en" json:"display_name_en"`
	Description   *string `bun:"description" json:"description"`
	IconName      *string `bun:"icon_name" json:"icon_name"`
	ColorHex      *string `bun:"color_hex" json:"color_hex"`

	// Hierarchy and ordering
	Level     int `bun:"level,default:0" json:"level"`
	SortOrder int `bun:"sort_order,default:0" json:"sort_order"`

	// Matching keywords for auto-categorization
	Keywords         []string `bun:"keywords,array" json:"keywords"`
	ExcludedKeywords []string `bun:"excluded_keywords,array" json:"excluded_keywords"`

	// Usage statistics
	ProductCount int `bun:"product_count,default:0" json:"product_count"`
	UsageCount   int `bun:"usage_count,default:0" json:"usage_count"`

	// Status
	IsActive bool `bun:"is_active,default:true" json:"is_active"`

	CreatedAt time.Time `bun:"created_at,default:now()" json:"created_at"`
	UpdatedAt time.Time `bun:"updated_at,default:now()" json:"updated_at"`

	// Relations
	Parent   *ProductCategory   `bun:"rel:belongs-to,join:parent_id=id" json:"parent,omitempty"`
	Children []*ProductCategory `bun:"rel:has-many,join:id=parent_id" json:"children,omitempty"`
}

// ProductTag represents a flexible tag system for products
type ProductTag struct {
	bun.BaseModel `bun:"table:product_tags,alias:pt"`

	ID             int    `bun:"id,pk,autoincrement" json:"id"`
	Name           string `bun:"name,unique,notnull" json:"name"`
	NormalizedName string `bun:"normalized_name,notnull" json:"normalized_name"`

	// Display information
	DisplayNameLT string  `bun:"display_name_lt,notnull" json:"display_name_lt"`
	DisplayNameEN *string `bun:"display_name_en" json:"display_name_en"`
	Description   *string `bun:"description" json:"description"`
	ColorHex      *string `bun:"color_hex" json:"color_hex"`

	// Tag type and behavior
	TagType     string `bun:"tag_type,default:'general'" json:"tag_type"`
	IsSystemTag bool   `bun:"is_system_tag,default:false" json:"is_system_tag"`

	// Usage statistics
	UsageCount int `bun:"usage_count,default:0" json:"usage_count"`

	// Status
	IsActive bool `bun:"is_active,default:true" json:"is_active"`

	CreatedAt time.Time `bun:"created_at,default:now()" json:"created_at"`
	UpdatedAt time.Time `bun:"updated_at,default:now()" json:"updated_at"`
}

// ShoppingListCategory represents custom categories within a shopping list
type ShoppingListCategory struct {
	bun.BaseModel `bun:"table:shopping_list_categories,alias:slc"`

	ID             int64     `bun:"id,pk,autoincrement" json:"id"`
	ShoppingListID int64     `bun:"shopping_list_id,notnull" json:"shopping_list_id"`
	UserID         uuid.UUID `bun:"user_id,notnull" json:"user_id"`
	Name           string    `bun:"name,notnull" json:"name"`
	ColorHex       *string   `bun:"color_hex" json:"color_hex"`
	IconName       *string   `bun:"icon_name" json:"icon_name"`
	SortOrder      int       `bun:"sort_order,default:0" json:"sort_order"`
	ItemCount      int       `bun:"item_count,default:0" json:"item_count"`

	CreatedAt time.Time `bun:"created_at,default:now()" json:"created_at"`

	// Relations
	ShoppingList *ShoppingList `bun:"rel:belongs-to,join:shopping_list_id=id" json:"shopping_list,omitempty"`
	User         *User         `bun:"rel:belongs-to,join:user_id=id" json:"user,omitempty"`
}

// UserTag represents user's personal tag preferences
type UserTag struct {
	bun.BaseModel `bun:"table:user_tags,alias:ut"`

	ID          int64      `bun:"id,pk,autoincrement" json:"id"`
	UserID      uuid.UUID  `bun:"user_id,notnull" json:"user_id"`
	TagName     string     `bun:"tag_name,notnull" json:"tag_name"`
	DisplayName *string    `bun:"display_name" json:"display_name"`
	ColorHex    *string    `bun:"color_hex" json:"color_hex"`
	UsageCount  int        `bun:"usage_count,default:0" json:"usage_count"`
	LastUsedAt  *time.Time `bun:"last_used_at" json:"last_used_at"`

	CreatedAt time.Time `bun:"created_at,default:now()" json:"created_at"`

	// Relations
	User *User `bun:"rel:belongs-to,join:user_id=id" json:"user,omitempty"`
}

// TagType represents different types of tags
type TagType string

const (
	TagTypeGeneral  TagType = "general"
	TagTypeDietary  TagType = "dietary"
	TagTypeAllergen TagType = "allergen"
	TagTypeQuality  TagType = "quality"
	TagTypeSeasonal TagType = "seasonal"
)

// IsRootCategory returns true if this category has no parent
func (pc *ProductCategory) IsRootCategory() bool {
	return pc.ParentID == nil
}

// GetFullPath returns the full category path (e.g., "Food > Dairy > Milk")
func (pc *ProductCategory) GetFullPath() string {
	if pc.Parent == nil {
		return pc.DisplayNameLT
	}
	return pc.Parent.GetFullPath() + " > " + pc.DisplayNameLT
}

// AddKeyword adds a keyword for category matching
func (pc *ProductCategory) AddKeyword(keyword string) {
	for _, existing := range pc.Keywords {
		if existing == keyword {
			return
		}
	}
	pc.Keywords = append(pc.Keywords, keyword)
	pc.UpdatedAt = time.Now()
}

// AddExcludedKeyword adds a keyword that excludes items from this category
func (pc *ProductCategory) AddExcludedKeyword(keyword string) {
	for _, existing := range pc.ExcludedKeywords {
		if existing == keyword {
			return
		}
	}
	pc.ExcludedKeywords = append(pc.ExcludedKeywords, keyword)
	pc.UpdatedAt = time.Now()
}

// IncrementUsage increments the usage counter
func (pc *ProductCategory) IncrementUsage() {
	pc.UsageCount++
	pc.UpdatedAt = time.Now()
}

// IsSystemTagType returns true if this is a system-defined tag
func (pt *ProductTag) IsSystemTagType() bool {
	return pt.IsSystemTag
}

// IncrementUsage increments the tag usage counter
func (pt *ProductTag) IncrementUsage() {
	pt.UsageCount++
	pt.UpdatedAt = time.Now()
}

// IsDietaryTag returns true if this tag is related to dietary restrictions
func (pt *ProductTag) IsDietaryTag() bool {
	return pt.TagType == string(TagTypeDietary)
}

// IsAllergenTag returns true if this tag represents an allergen
func (pt *ProductTag) IsAllergenTag() bool {
	return pt.TagType == string(TagTypeAllergen)
}

// UpdateUsage updates the user tag usage statistics
func (ut *UserTag) UpdateUsage() {
	ut.UsageCount++
	now := time.Now()
	ut.LastUsedAt = &now
}

// GetDisplayName returns the display name or tag name if display name is not set
func (ut *UserTag) GetDisplayName() string {
	if ut.DisplayName != nil {
		return *ut.DisplayName
	}
	return ut.TagName
}

// TableName returns the table name for ProductCategory
func (pc *ProductCategory) TableName() string {
	return "product_categories"
}

// TableName returns the table name for ProductTag
func (pt *ProductTag) TableName() string {
	return "product_tags"
}

// TableName returns the table name for ShoppingListCategory
func (slc *ShoppingListCategory) TableName() string {
	return "shopping_list_categories"
}

// TableName returns the table name for UserTag
func (ut *UserTag) TableName() string {
	return "user_tags"
}
