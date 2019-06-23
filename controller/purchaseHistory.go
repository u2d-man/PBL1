package controller

import (
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/PBL1/model"
	"github.com/PBL1/service"
	"github.com/gin-gonic/gin"
)

// GetPurchaseHistoriesByUserID ...ユーザーIDを元に購入履歴を取得する
func GetPurchaseHistoriesByUserID(c *gin.Context) {
	purchaseHistory := PurchaseHistory{}
	purchaseHistories := []PurchaseHistory{}
	purchaseHistoryCard := PurchaseHistoryCard{}
	purchaseHistoryCards := []PurchaseHistoryCard{}
	foods := []Food{}
	food := Food{}
	modelRecipePurchaseHistories := []model.RecipePurchaseHistory{}
	modelFoodPurchaseHistories := []model.FoodPurchaseHistory{}
	purchaseDates := []time.Time{}
	recipe := model.Recipe{}
	var err error

	userID := c.Param("user_id")

	// errCode := AuthCheck(c, userID)
	// if errCode != nil {
	// 	c.AbortWithStatus(http.StatusUnauthorized)
	// 	return
	// }

	//特定ユーザーの購入日データ群を取得
	purchaseDates, err = service.GetPurchaseDatesByUserID(userID)
	if err != nil {
		log.Println("登録されていないユーザIDです")
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	for _, purchaseDate := range purchaseDates {
		purchaseHistory.Date = purchaseDate
		modelRecipePurchaseHistories, err = service.GetmodelRecipePurchaseHistoriesByUserIDAndPurchaseDate(userID, purchaseDate)

		for _, modelRecipePurchaseHistory := range modelRecipePurchaseHistories {
			// 材料の中身をリセット
			foods = nil

			// レシピの値段・ポイント・個数を挿入
			purchaseHistoryCard.Price = modelRecipePurchaseHistory.Price
			purchaseHistoryCard.Point = modelRecipePurchaseHistory.Point
			purchaseHistoryCard.RecipeCount = *modelRecipePurchaseHistory.RecipeCount

			// レシピデータを取得
			recipe = service.GetRecipeByRecipeID(modelRecipePurchaseHistory.RecipeID)

			// レシピの名前・イメージ画像URLを挿入
			purchaseHistoryCard.RecipeName = recipe.Name
			purchaseHistoryCard.RecipeImageURL = recipe.ImageURL

			// 値段とポイントの合計値を計算
			purchaseHistory.TotalPrice += purchaseHistoryCard.Price
			purchaseHistory.TotalPoint += purchaseHistoryCard.Point

			// レシピごとの食材購入履歴データを取得
			modelFoodPurchaseHistories, err = service.GetFoodPurchaseHistoriesByUserIDAndRecipeID(userID, recipe.ID)
			if err != nil {
				c.AbortWithStatus(http.StatusBadRequest)
				return
			}

			for _, modelFoodPurchaseHistory := range modelFoodPurchaseHistories {
				food.ID = modelFoodPurchaseHistory.FoodID
				food.FoodCount = *modelFoodPurchaseHistory.FoodCount
				food.Name, err = service.GetFoodNameByID(modelFoodPurchaseHistory.FoodID)

				if modelFoodPurchaseHistory.Quantity == 0 {
					food.Quantity = modelFoodPurchaseHistory.Unit
				} else {
					food.Quantity = strconv.FormatUint(uint64(modelFoodPurchaseHistory.Quantity), 10) + modelFoodPurchaseHistory.Unit
				}
				foods = append(foods, food)
			}

			purchaseHistoryCard.Foods = foods

			purchaseHistoryCards = append(purchaseHistoryCards, purchaseHistoryCard)
		}
		purchaseHistory.PurchaseHistoryCards = purchaseHistoryCards
	}
	purchaseHistories = append(purchaseHistories, purchaseHistory)

	c.JSON(http.StatusOK, purchaseHistories)

}

// PostPurchaseHistoriesByUserID ...ユーザーごとのカート内データを購入履歴に追加する
func PostPurchaseHistoriesByUserID(c *gin.Context) {
	stringUserID := c.PostForm("user_id")
	// carts := []model.Cart{}
	uintRecipeID := uint(1)

	// ユーザー毎のカートに入っているrecipeIDを取ってくる
	recipeIDs, _ :=  service.GetRecipeIDsInCartByUserID(stringUserID)
	log.Printf("recipeIDs :", recipeIDs)

	errCode := AuthCheck(c, stringUserID)
	if errCode != nil {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	// ユーザー毎のカートに入っているrecipeの分だけloop
	for _, recipeID := range recipeIDs {
		uintRecipeID = uint(recipeID)
		log.Printf("uintRecipeID : ", uintRecipeID, recipeID)
		// recipeIDを引数で渡し、該当するデータのPrice, Point取得
		recipePrice, recipePoint, _ := service.GetRecipePriceAndPointByID(recipeIDs)
		// ユーザー毎のカートに入っているrecipeセット数を取得
		recipeSetCountInCart, _ := service.GetRecipeSetCountInCartsByUserID(stringUserID)
		log.Printf("recipeSetCountInCart :", recipeSetCountInCart)
		// ユーザー毎のカートに入っているrecipeのfoodIDを取得
		foodIDsInCarts, _ := service.GetFoodIDsInCartByUserID(stringUserID, uintRecipeID)
		log.Printf("foodIDsInCarts", foodIDsInCarts)

		for _, foodIDsInCart := range foodIDsInCarts {
			// ユーザー毎のカートに入っているrecipeIDとfoodIDから材料の詳細を取得
			ingredientsUserCarts, _ := service.GetIngredientsByRecipeIDAndFoodID(uintRecipeID, foodIDsInCart.FoodID)
			// foodPurchaseHistoryにinsert
			service.InsertFoodCartContentsToPuchaseHistory(stringUserID, foodIDsInCarts, ingredientsUserCarts)
			// recipePurchaseHistoryにinsert
			service.InsertRecipeCartContentsToPuchaseHistory(stringUserID, recipeSetCountInCart, recipePrice, recipePoint)
		}
	}
 	c.AbortWithStatus(http.StatusOK)
}
