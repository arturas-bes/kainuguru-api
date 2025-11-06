package resolvers

import (
	"github.com/kainuguru/kainuguru-api/internal/graphql/generated"
)

// This file contains only the interface bindings for gqlgen.
// All resolver implementations are in separate files (flyer.go, product.go, query.go, store.go, etc.)
// DO NOT add resolver method implementations here - they belong in their respective files.

// Flyer returns generated.FlyerResolver implementation.
func (r *Resolver) Flyer() generated.FlyerResolver { return &flyerResolver{r} }

// FlyerPage returns generated.FlyerPageResolver implementation.
func (r *Resolver) FlyerPage() generated.FlyerPageResolver { return &flyerPageResolver{r} }

// Mutation returns generated.MutationResolver implementation.
func (r *Resolver) Mutation() generated.MutationResolver { return &mutationResolver{r} }

// PriceAlert returns generated.PriceAlertResolver implementation.
func (r *Resolver) PriceAlert() generated.PriceAlertResolver { return &priceAlertResolver{r} }

// PriceHistory returns generated.PriceHistoryResolver implementation.
func (r *Resolver) PriceHistory() generated.PriceHistoryResolver { return &priceHistoryResolver{r} }

// Product returns generated.ProductResolver implementation.
func (r *Resolver) Product() generated.ProductResolver { return &productResolver{r} }

// ProductMaster returns generated.ProductMasterResolver implementation.
func (r *Resolver) ProductMaster() generated.ProductMasterResolver { return &productMasterResolver{r} }

// Query returns generated.QueryResolver implementation.
func (r *Resolver) Query() generated.QueryResolver { return &queryResolver{r} }

// ShoppingList returns generated.ShoppingListResolver implementation.
func (r *Resolver) ShoppingList() generated.ShoppingListResolver { return &shoppingListResolver{r} }

// ShoppingListCategory returns generated.ShoppingListCategoryResolver implementation.
func (r *Resolver) ShoppingListCategory() generated.ShoppingListCategoryResolver {
	return &shoppingListCategoryResolver{r}
}

// ShoppingListItem returns generated.ShoppingListItemResolver implementation.
func (r *Resolver) ShoppingListItem() generated.ShoppingListItemResolver {
	return &shoppingListItemResolver{r}
}

// Store returns generated.StoreResolver implementation.
func (r *Resolver) Store() generated.StoreResolver { return &storeResolver{r} }

// User returns generated.UserResolver implementation.
func (r *Resolver) User() generated.UserResolver { return &userResolver{r} }

// Resolver type definitions - these embed *Resolver to access services
type flyerResolver struct{ *Resolver }
type flyerPageResolver struct{ *Resolver }
type mutationResolver struct{ *Resolver }
type priceAlertResolver struct{ *Resolver }
type priceHistoryResolver struct{ *Resolver }
type productResolver struct{ *Resolver }
type productMasterResolver struct{ *Resolver }
type queryResolver struct{ *Resolver }
type shoppingListResolver struct{ *Resolver }
type shoppingListCategoryResolver struct{ *Resolver }
type shoppingListItemResolver struct{ *Resolver }
type storeResolver struct{ *Resolver }
type userResolver struct{ *Resolver }
