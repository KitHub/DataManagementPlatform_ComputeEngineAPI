package service

import (
	"context"
	"sync"

	"github.com/KitHub/DataManagementPlatform_ComputeEngineAPI/logic"
	computeEngineAPIProtocol "github.com/KitHub/protocols/DataManagementPlatform_ComputeEngineAPI"
)

var (
	computeEngineServiceInstance *ComputeEngineService
	computeEngineServiceOnce     sync.Once
)

type ComputeEngineService struct {
	computeEngineAPIProtocol.UnimplementedComputeEngineAPIServer
	packageLogic *logic.PackageLogic
}

// GetPackageById implements [DataManagementPlatform_ComputeEngineAPI.ComputeEngineAPIServer].
func (c *ComputeEngineService) GetPackageById(context.Context, *computeEngineAPIProtocol.GetPackageByIdRequest) (*computeEngineAPIProtocol.GetPackageByIdResponse, error) {
	panic("unimplemented")
}

// GetPackageByName implements [DataManagementPlatform_ComputeEngineAPI.ComputeEngineAPIServer].
func (c *ComputeEngineService) GetPackageByName(context.Context, *computeEngineAPIProtocol.GetPackageByNameRequest) (*computeEngineAPIProtocol.GetPackageByNameResponse, error) {
	panic("unimplemented")
}

// GetPackageListASCByLastId implements [DataManagementPlatform_ComputeEngineAPI.ComputeEngineAPIServer].
func (c *ComputeEngineService) GetPackageListASCByLastId(context.Context, *computeEngineAPIProtocol.GetPackageListASCByLastIdRequest) (*computeEngineAPIProtocol.GetPackageListASCByLastIdResponse, error) {
	panic("unimplemented")
}

// RegisterPackage implements [DataManagementPlatform_ComputeEngineAPI.ComputeEngineAPIServer].
func (c *ComputeEngineService) RegisterPackage(context.Context, *computeEngineAPIProtocol.RegisterPackageRequest) (*computeEngineAPIProtocol.RegisterPackageResponse, error) {
	panic("unimplemented")
}

func NewComputeEngineService(packageLogic *logic.PackageLogic) *ComputeEngineService {
	computeEngineServiceOnce.Do(func() {
		computeEngineServiceInstance = &ComputeEngineService{
			packageLogic: packageLogic,
		}
	})
	return computeEngineServiceInstance
}
