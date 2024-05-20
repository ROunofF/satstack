package handlers

import (
	"net/http"
	"sort"
	"strconv"
	"strings"

	"github.com/ledgerhq/satstack/httpd/svc"
	"github.com/ledgerhq/satstack/types"
	"github.com/ledgerhq/satstack/utils"

	"github.com/gin-gonic/gin"
)

func GetAddresses(s svc.AddressesService) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		addresses, shouldReturn := getAdresses(ctx, s)
		if shouldReturn {
			return
		}

		ctx.JSON(http.StatusOK, addresses)
	}
}

func getAdresses(ctx *gin.Context, s svc.AddressesService) (types.Addresses, bool) {
	param := ctx.Param("addresses")
	blockHashQuery := ctx.Query("block_hash")
	blockHeightQuery := ctx.Query("block_height")

	addressList := strings.Split(param, ",")

	var blockHash *string
	if blockHashQuery != "" {
		blockHash = &blockHashQuery
	}

	var blockHeight *int32
	if blockHeightQuery != "" {
		n, _ := strconv.ParseInt(blockHeightQuery, 10, 32)
		i32 := int32(n)
		blockHeight = &i32
	}

	addresses, err := s.GetAddresses(addressList, blockHash, blockHeight)
	if err != nil {
		ctx.String(http.StatusNotFound, "text/plain", []byte(err.Error()))
		return types.Addresses{}, true
	}

	// FIXME: libcore relies on the order of the transactions, in order to
	//        correctly compute operation values (aka amounts). This order
	//        appears to be based on the ReceivedAt field, although it is
	//        not documented in the Ledger BE project.
	//
	//        The bug seems to manifest itself only on accounts with a
	//        large number of operations.
	sort.Slice(addresses.Transactions[:], func(i, j int) bool {
		iReceivedAt, iErr := utils.ParseRFC3339Timestamp(addresses.Transactions[i].ReceivedAt)
		jReceivedAt, jErr := utils.ParseRFC3339Timestamp(addresses.Transactions[j].ReceivedAt)

		if iErr != nil || jErr != nil {
			// Still a semi-reliable way of comparing RFC3339 timestamps.
			return addresses.Transactions[i].ReceivedAt < addresses.Transactions[j].ReceivedAt
		}

		return *iReceivedAt < *jReceivedAt
	})
	return addresses, false
}

func GetV4Addresses(s svc.AddressesService) gin.HandlerFunc {
	return func(ctx *gin.Context) {

		addresses, shouldReturn := getAdresses(ctx, s)
		if shouldReturn {
			return
		}
		type resp struct {
			Data  []types.Transaction `json:"data"`
			Token *string             `json:"token"`
		}

		response := resp{Data: addresses.Transactions}
		ctx.JSON(http.StatusOK, response)
	}
}
