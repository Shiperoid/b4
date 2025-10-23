package handler

import "github.com/daniellavrushin/b4/nfq"

var (
	globalPool *nfq.Pool
)

func SetNFQPool(pool *nfq.Pool) {
	globalPool = pool
}
