package grpc

import (
	"context"

	pb "omnistock-ai/proto/inventory/v1"
	"omnistock-ai/services/inventory-service/internal/domain"
)

// InventoryGrpcHandler handles incoming gRPC requests for the inventory service
type InventoryGrpcHandler struct {
	pb.UnimplementedInventoryServiceServer
	usecase domain.ProductUsecase
}

// NewInventoryGrpcHandler creates a new instance of the gRPC handler
func NewInventoryGrpcHandler(us domain.ProductUsecase) *InventoryGrpcHandler {
	return &InventoryGrpcHandler{
		usecase: us,
	}
}

// ReserveStock is called by the OMS service via gRPC to reserve items
func (h *InventoryGrpcHandler) ReserveStock(ctx context.Context, req *pb.ReserveStockRequest) (*pb.ReserveStockResponse, error) {

	// 1. Call the Usecase to perform actual database update
	err := h.usecase.ReserveStock(req.ProductSku, int(req.Quantity))

	// 2. Return false with error message if out of stock or not found
	if err != nil {
		return &pb.ReserveStockResponse{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	// 3. Return true if database update is successful
	return &pb.ReserveStockResponse{
		Success: true,
		Message: "Stock reserved successfully for SKU: " + req.ProductSku,
	}, nil
}
