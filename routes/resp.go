package routes

type NormalOutput struct {
	Body struct {
		Message string `json:"message" example:"Success" doc:"Status message."`
	}
}
