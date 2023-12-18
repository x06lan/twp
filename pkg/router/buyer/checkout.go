package buyer

import (
	"encoding/json"
	"math"
	"net/http"
	"sort"

	"github.com/jykuo-love-shiritori/twp/db"
	"github.com/jykuo-love-shiritori/twp/pkg/constants"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

type couponDiscount struct {
	ID            int32          `json:"id"`
	Name          string         `json:"name"`
	Type          db.CouponType  `json:"type"`
	Scope         db.CouponScope `json:"scope"`
	Description   string         `json:"description"`
	Discount      float64        `json:"discount"`
	DiscountValue int32          `json:"discount_value"`
}

type checkout struct {
	Subtotal      int32            `json:"subtotal"`
	Shipment      int32            `json:"shipment"`
	TotalDiscount int32            `json:"total_discount"`
	Coupons       []couponDiscount `json:"coupons"`
	Total         int32            `json:"total"`
	Payments      json.RawMessage  `json:"payments"`
}

func getShipmentFee(total int32) int32 {
	return int32(math.Log(float64(total)))
}
func getDiscountValue(price float64, discount float64, couponType db.CouponType) int32 {
	switch couponType {
	case db.CouponTypePercentage:
		return int32(price * discount / 100)
	case db.CouponTypeFixed:
		return min(int32(discount), int32(price))
	}
	return 0
}

// @Summary		Buyer Get Checkout
// @Description	Get all checkout data
// @Tags			Buyer, Checkout
// @Produce		json
// @Param			cart_id	path		int	true	"Cart ID"
// @Success		200		{object}	checkout
// @Failure		400		{object}	echo.HTTPError
// @Failure		500		{object}	echo.HTTPError
// @Router			/buyer/cart/{cart_id}/checkout [get]
func GetCheckout(pg *db.DB, logger *zap.SugaredLogger) echo.HandlerFunc {
	return func(c echo.Context) error {
		username := "Buyer"
		result := checkout{Coupons: []couponDiscount{}}
		var cartID int32
		if err := echo.PathParamsBinder(c).Int32("cart_id", &cartID).BindError(); err != nil {
			logger.Errorw("failed to parse cart_id", "error", err)
			return echo.NewHTTPError(http.StatusBadRequest)
		}
		// this will validate product stock and cart legitimacy
		valid, err := pg.Queries.ValidateProductsInCart(c.Request().Context(), db.ValidateProductsInCartParams{
			Username: username,
			CartID:   cartID,
		})
		if err != nil {
			logger.Errorw("failed to validate product in cart", "error", err)
			return echo.NewHTTPError(http.StatusInternalServerError)
		} else if !valid {
			return echo.NewHTTPError(http.StatusBadRequest, "Some product is not available now")
		}
		productCount := make(map[int32]int32)
		productTag := make(map[int32]*tagSet)
		products, err := pg.Queries.GetProductFromCart(c.Request().Context(), cartID)
		if err != nil {
			logger.Errorw("failed to get product from cart", "error", err)
			return echo.NewHTTPError(http.StatusInternalServerError)
		}
		sort.Slice(products, func(i, j int) bool {
			return products[i].Price.Int.Cmp(products[j].Price.Int) > 0
		})
		for _, product := range products {
			productCount[product.ProductID] = product.Quantity
			price, err := product.Price.Float64Value()
			if err != nil {
				logger.Errorw("failed to get price", "error", err)
				return echo.NewHTTPError(http.StatusInternalServerError)
			}
			result.Subtotal += int32(price.Float64 * float64(product.Quantity))
			tags, err := pg.Queries.GetProductTag(c.Request().Context(), product.ProductID)
			if err != nil {
				logger.Errorw("failed to get product tag", "error", err)
				return echo.NewHTTPError(http.StatusInternalServerError)
			}
			productTag[product.ProductID] = NewTagSet(tags)
		}
		result.Shipment = getShipmentFee(result.Subtotal)

		var params db.GetCouponsFromCartParams
		params.CartID = cartID
		params.Username = username
		coupons, err := pg.Queries.GetCouponsFromCart(c.Request().Context(), params)
		if err != nil {
			logger.Errorw("failed to get coupons from cart", "error", err)
			return echo.NewHTTPError(http.StatusInternalServerError)
		}

		sort.Slice(coupons, func(i, j int) bool {
			if coupons[i].Type == coupons[j].Type {
				return coupons[i].Discount.Int.Cmp(coupons[j].Discount.Int) > 0
			}
			return coupons[i].Type < coupons[j].Type
		})

		totalDiscount := int32(0)
		for _, coupon := range coupons {
			var cp couponDiscount = couponDiscount{
				ID:          coupon.ID,
				Name:        coupon.Name,
				Type:        coupon.Type,
				Scope:       coupon.Scope,
				Description: coupon.Description,
			}
			discount, err := coupon.Discount.Float64Value()
			if err != nil {
				logger.Errorw("failed to get discount", "error", err)
				return echo.NewHTTPError(http.StatusInternalServerError)
			}
			cp.Discount = discount.Float64
			if cp.Type == db.CouponTypeShipping {
				cp.DiscountValue = result.Shipment * (int32(cp.Discount / 100))
				totalDiscount += cp.DiscountValue
				result.Coupons = append(result.Coupons, cp)
				continue
			}
			tags, err := pg.Queries.GetCouponTag(c.Request().Context(), coupon.ID)
			if err != nil {
				logger.Errorw("failed to get coupon tag", "error", err)
				return echo.NewHTTPError(http.StatusInternalServerError)
			}
			couponTags := NewTagSet(tags)
			for _, product := range products {
				if productCount[product.ProductID] == 0 {
					continue
				}
				if productTag[product.ProductID].Intersect(couponTags) {
					productCount[product.ProductID] -= 1
					price, err := product.Price.Float64Value()
					if err != nil {
						logger.Errorw("failed to get price", "error", err)
						return echo.NewHTTPError(http.StatusInternalServerError)
					}
					logger.Infow("match coupon", "product_id", product.ProductID, "coupon_id", coupon.ID, "discount", cp.Discount, "price", price.Float64)
					cp.DiscountValue = getDiscountValue(price.Float64, cp.Discount, cp.Type)
				}
			}
			totalDiscount += cp.DiscountValue
			result.Coupons = append(result.Coupons, cp)
		}
		result.TotalDiscount = totalDiscount
		result.Total = max(0, result.Subtotal+result.Shipment-result.TotalDiscount)
		result.Payments, err = pg.Queries.GetCreditCard(c.Request().Context(), username)
		if err != nil {
			logger.Errorw("failed to get credit card", "error", err)
			return echo.NewHTTPError(http.StatusInternalServerError)
		}
		return c.JSON(http.StatusOK, result)
	}
}

type PaymentMethod struct {
	CreditCard json.RawMessage `swaggertype:"object"`
}

// @Summary		Buyer Checkout
// @Description	Checkout
// @Tags			Buyer, Checkout
// @Accept			json
// @Produce		json
// @param			cart_id			path		int				true	"Cart ID"
// @Param			payment_method	body		PaymentMethod	true	"Payment" Example
// @Success		200				{string}	string			constants.SUCCESS
// @Failure		400				{object}	echo.HTTPError
// @Failure		500				{object}	echo.HTTPError
// @Router			/buyer/cart/{cart_id}/checkout [post]
func Checkout(pg *db.DB, logger *zap.SugaredLogger) echo.HandlerFunc {
	return func(c echo.Context) error {
		username := "Buyer"
		var cartID int32
		if err := echo.PathParamsBinder(c).Int32("cart_id", &cartID).BindError(); err != nil {
			logger.Errorw("failed to parse cart_id", "error", err)
			return echo.NewHTTPError(http.StatusBadRequest)
		}
		var param db.ValidatePaymentParams
		param.Username = username
		if err := c.Bind(&param); err != nil {
			logger.Errorw("failed to bind payment", "error", err)
			return echo.NewHTTPError(http.StatusBadRequest)
		}
		if valid, err := pg.Queries.ValidatePayment(c.Request().Context(), param); err != nil {
			logger.Errorw("failed to validate payment", "error", err)
			return echo.NewHTTPError(http.StatusInternalServerError)
		} else if !valid {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid payment")
		}
		products, err := pg.Queries.GetProductFromCart(c.Request().Context(), cartID)
		if err != nil {
			logger.Errorw("failed to get product from cart", "error", err)
			return echo.NewHTTPError(http.StatusInternalServerError)
		}
		sort.Slice(products, func(i, j int) bool {
			return products[i].Price.Int.Cmp(products[j].Price.Int) > 0
		})
		subtotal := int32(0)
		productCount := make(map[int32]int32)
		productTag := make(map[int32]*tagSet)
		// calculate subtotal and get tags and counts
		for _, product := range products {
			err := pg.Queries.UpdateProductVersion(c.Request().Context(), product.ProductID)
			if err != nil {
				logger.Errorw("failed to update product version", "error", err)
				return echo.NewHTTPError(http.StatusInternalServerError)
			}
			price, err := product.Price.Float64Value()
			if err != nil {
				logger.Errorw("failed to get price", "error", err)
				return echo.NewHTTPError(http.StatusInternalServerError)
			}
			subtotal += int32(price.Float64 * float64(product.Quantity))
			productCount[product.ProductID] = product.Quantity
			tags, err := pg.Queries.GetProductTag(c.Request().Context(), product.ProductID)
			if err != nil {
				logger.Errorw("failed to get product tag", "error", err)
				return echo.NewHTTPError(http.StatusInternalServerError)
			}
			productTag[product.ProductID] = NewTagSet(tags)
		}
		// this will validate cart and product legitimacy
		if valid, err := pg.Queries.ValidateProductsInCart(c.Request().Context(), db.ValidateProductsInCartParams{
			Username: username,
			CartID:   cartID,
		}); err != nil {
			logger.Errorw("failed to validate product in cart", "error", err)
			return echo.NewHTTPError(http.StatusInternalServerError)
		} else if !valid {
			return echo.NewHTTPError(http.StatusBadRequest, "Some product is not available now")
		}
		shipment := getShipmentFee(int32(subtotal))
		var params db.GetCouponsFromCartParams
		params.CartID = cartID
		params.Username = username
		coupons, err := pg.Queries.GetCouponsFromCart(c.Request().Context(), params)
		if err != nil {
			logger.Errorw("failed to get coupons from cart", "error", err)
			return echo.NewHTTPError(http.StatusInternalServerError)
		}
		// sort to make customer get most discount
		sort.Slice(coupons, func(i, j int) bool {
			if coupons[i].Type == coupons[j].Type {
				return coupons[i].Discount.Int.Cmp(coupons[j].Discount.Int) > 0
			}
			return coupons[i].Type < coupons[j].Type
		})

		totalDiscount := int32(0)
		// match coupon with product
		for _, coupon := range coupons {
			discount, err := coupon.Discount.Float64Value()
			if err != nil {
				logger.Errorw("failed to get discount", "error", err)
				return echo.NewHTTPError(http.StatusInternalServerError)
			}
			dc := discount.Float64
			if coupon.Type == db.CouponTypeShipping {
				totalDiscount += shipment * (int32(dc / 100))
				continue
			}
			tags, err := pg.Queries.GetCouponTag(c.Request().Context(), coupon.ID)
			if err != nil {
				logger.Errorw("failed to get coupon tag", "error", err)
				return echo.NewHTTPError(http.StatusInternalServerError)
			}
			couponTags := NewTagSet(tags)
			for _, product := range products {
				if productCount[product.ProductID] == 0 {
					continue
				}
				if productTag[product.ProductID].Intersect(couponTags) {
					productCount[product.ProductID] -= 1
					price, err := product.Price.Float64Value()
					if err != nil {
						logger.Errorw("failed to get price", "error", err)
						return echo.NewHTTPError(http.StatusInternalServerError)
					}
					totalDiscount += getDiscountValue(price.Float64, dc, coupon.Type)
					break
				}
			}
		}
		total := max(0, int32(subtotal)+shipment-totalDiscount)
		// if total < 0, get achievement “There is nothing more expensive than something free”

		if err := pg.Queries.Checkout(c.Request().Context(),
			db.CheckoutParams{
				Username:   username,
				Shipment:   shipment,
				CartID:     cartID,
				TotalPrice: total}); err != nil {
			logger.Errorw("failed to checkout", "error", err)
			return echo.NewHTTPError(http.StatusInternalServerError)
		}
		return c.JSON(http.StatusOK, constants.SUCCESS)
	}
}
