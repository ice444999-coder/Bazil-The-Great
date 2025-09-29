package controllers

import (
	service "ares_api/internal/interfaces/service"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
)

type AssetController struct {
	Service service.AssetService
	LedgerService service.LedgerService
}

func NewAssetController(s service.AssetService , l service.LedgerService) *AssetController {
	return &AssetController{Service: s , LedgerService: l}
}

//GetAllCoins godoc
// @Summary      Get all coins
// @Description  Fetch the list of coins from CoinGecko with optional limit
// @Tags         coins
// @Produce      json
// @Param        limit  query     int  false  "Limit number of coins to return"  default(100)
// @Success      200  {array}  dto.CoinDTO
// @Failure      500  {object}  map[string]string
// @Router        /assets/coins [get]
func (c *AssetController) GetAllCoins(ctx *gin.Context) {
	limitStr := ctx.DefaultQuery("limit", "100")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 100
	}

	coins, err := c.Service.GetAllCoins(limit)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	_ = c.LedgerService.Append(0,  "GetAllCoins", "Fetched list of coins with limit: " + limitStr)
	ctx.JSON(http.StatusOK, coins)
}

// GetCoinMarket godoc
// @Summary      Get coin market data
// @Description  Fetch real-time market data for a specific coin
// @Tags         coins
// @Produce      json
// @Param        id           path      string  true   "Coin ID"
// @Param        vs_currency  query     string  false  "Currency to fetch prices in (default: USD)"
// @Success      200  {object}  dto.CoinMarketDTO
// @Failure      500  {object}  map[string]string
// @Router       /assets/coins/{id}/market [get]
func (c *AssetController) GetCoinMarket(ctx *gin.Context) {
    id := ctx.Param("id")
    currency := ctx.DefaultQuery("vs_currency", "usd") // default to USD
    coin, err := c.Service.GetCoinMarket(id, currency)
    if err != nil {
        ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
	_ = c.LedgerService.Append(0,  "GetCoinMarket", "Fetched market data for coin ID: " + id)
    ctx.JSON(http.StatusOK, coin)
}


// GetTopMovers godoc
// @Summary      Get top movers
// @Description  Fetch top N coins by 24h price change
// @Tags         coins
// @Produce      json
// @Param        limit  query     int  false  "Number of top movers"  default(10)
// @Success      200    {array}   dto.TopMoverDTO
// @Failure      500    {object}  map[string]string
// @Router      /assets/coins/top-movers [get]
func (c *AssetController) GetTopMovers(ctx *gin.Context) {
	limitStr := ctx.DefaultQuery("limit", "10")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 10
	}

	movers, err := c.Service.GetTopMovers(limit)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	_ = c.LedgerService.Append(0,  "GetTopMovers", "Fetched top movers with limit: " + limitStr)
	ctx.JSON(http.StatusOK, movers)
}


// GetSupportedVSCurrencies godoc
// @Summary      Get all supported vs_currency options
// @Description  Returns a list of supported currencies for market data
// @Tags         coins
// @Produce      json
// @Success      200  {array}  string
// @Failure      500  {object}  map[string]string
// @Router       /assets/vs_currencies [get]
func (c *AssetController) GetSupportedVSCurrencies(ctx *gin.Context) {
    currencies, err := c.Service.GetSupportedVSCurrencies()
    if err != nil {
        ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
	_ = c.LedgerService.Append(0,  "GetSupportedVSCurrencies", "Fetched supported vs_currencies")
    ctx.JSON(http.StatusOK, currencies)
}
