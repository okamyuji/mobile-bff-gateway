package mobile

// HomeResponse はモバイルホーム画面向けに集約した小さいJSONです。
type HomeResponse struct {
	User          UserSummary           `json:"user"`
	PaymentOrders []PaymentOrderSummary `json:"paymentOrders"`
	Account       AccountSummary        `json:"account"`
	Meta          ResponseMetadata      `json:"meta"`
}

// UserSummary は画面表示に必要なユーザー情報だけを持ちます。
type UserSummary struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// PaymentOrderSummary は画面表示に必要な決済注文情報だけを持ちます。
type PaymentOrderSummary struct {
	ID     string `json:"id"`
	Status string `json:"status"`
	Amount int    `json:"amount"`
}

// AccountSummary は画面表示に必要な口座情報だけを持ちます。
type AccountSummary struct {
	AvailableBalance int    `json:"availableBalance"`
	Currency         string `json:"currency"`
}

// ResponseMetadata はGateway側で付与する補助情報です。
type ResponseMetadata struct {
	Cached bool `json:"cached"`
}
