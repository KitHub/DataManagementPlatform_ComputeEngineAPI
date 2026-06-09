package service

import (
	"context"
	"log/slog"
	"sync"

	"github.com/KitHub/DataManagementPlatform_ComputeEngineAPI/entity"
	"github.com/KitHub/DataManagementPlatform_ComputeEngineAPI/logic"
	computeEngineAPIProtocol "github.com/KitHub/protocols/DataManagementPlatform_ComputeEngineAPI"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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
func (c *ComputeEngineService) GetPackageById(ctx context.Context, req *computeEngineAPIProtocol.GetPackageByIdRequest) (rsp *computeEngineAPIProtocol.GetPackageByIdResponse, err error) {
	slog.InfoContext(ctx, "GetPackageById", slog.Any("package_id", req.GetPackageId()))
	err = req.Validate()
	if err != nil {
		errMsg := "invalid request parameters: " + err.Error()
		slog.ErrorContext(ctx, errMsg, slog.Any("request", req.String()))
		return nil, status.Errorf(codes.InvalidArgument, "invalid request parameters")
	}

	packageEntity, err := c.packageLogic.QueryPackageById(ctx, req.GetPackageId())
	if err != nil {
		slog.ErrorContext(ctx, "queryPackageById failed", slog.Any("err", err))
		return nil, status.Errorf(codes.Internal, "server error")
	}

	rsp = &computeEngineAPIProtocol.GetPackageByIdResponse{
		ErrCode: 0,
		ErrMsg:  "ok",
		Data: &computeEngineAPIProtocol.GetPackageByIdResponseData{
			PackageInfo: convertPackageEntityToBasicPackageInfo(ctx, packageEntity),
		},
	}

	slog.InfoContext(ctx, "GetPackageById", slog.Any("req", req), slog.Any("rsp", rsp))
	return rsp, nil
}

// GetPackageByName implements [DataManagementPlatform_ComputeEngineAPI.ComputeEngineAPIServer].
func (c *ComputeEngineService) GetPackageByName(ctx context.Context, req *computeEngineAPIProtocol.GetPackageByNameRequest) (rsp *computeEngineAPIProtocol.GetPackageByNameResponse, err error) {
	slog.InfoContext(ctx, "GetPackageByName", slog.Any("package_name", req.GetName()))
	err = req.Validate()
	if err != nil {
		errMsg := "invalid request parameters: " + err.Error()
		slog.ErrorContext(ctx, errMsg, slog.Any("request", req.String()))
		return nil, status.Errorf(codes.InvalidArgument, "invalid request parameters")
	}

	packageEntity, err := c.packageLogic.QueryPackageByName(ctx, req.GetName())
	if err != nil {
		slog.ErrorContext(ctx, "queryPackageById failed", slog.Any("err", err))
		return nil, status.Errorf(codes.Internal, "server error")
	}

	rsp = &computeEngineAPIProtocol.GetPackageByNameResponse{
		ErrCode: 0,
		ErrMsg:  "ok",
		Data: &computeEngineAPIProtocol.GetPackageByNameResponseData{
			PackageInfo: convertPackageEntityToBasicPackageInfo(ctx, packageEntity),
		},
	}

	slog.InfoContext(ctx, "GetPackageById", slog.Any("req", req), slog.Any("rsp", rsp))
	return rsp, nil
}

// GetPackageListASCByLastId implements [DataManagementPlatform_ComputeEngineAPI.ComputeEngineAPIServer].
func (c *ComputeEngineService) GetPackageListASCByLastId(ctx context.Context, req *computeEngineAPIProtocol.GetPackageListASCByLastIdRequest) (rsp *computeEngineAPIProtocol.GetPackageListASCByLastIdResponse, err error) {
	slog.InfoContext(ctx, "GetPackageListASCByLastId", slog.Any("last_id", req.GetLastId()), slog.Any("limit", req.GetLimit()))
	err = req.Validate()
	if err != nil {
		errMsg := "invalid request parameters: " + err.Error()
		slog.ErrorContext(ctx, errMsg, slog.Any("request", req.String()))
		return nil, status.Errorf(codes.InvalidArgument, "invalid request parameters")
	}

	packageEntityList, err := c.packageLogic.QueryPackageListASCById(ctx, req.GetLastId(), req.GetLimit())
	if err != nil {
		slog.ErrorContext(ctx, "QueryPackageListASCById failed", slog.Any("err", err))
		return nil, status.Errorf(codes.Internal, "server error")
	}

	packageInfoList := make([]*computeEngineAPIProtocol.BasicPackageInfo, 0)
	for _, tmpPackageEntity := range packageEntityList {
		packageInfoList = append(packageInfoList, convertPackageEntityToBasicPackageInfo(ctx, tmpPackageEntity))
	}

	rsp = &computeEngineAPIProtocol.GetPackageListASCByLastIdResponse{
		ErrCode: 0,
		ErrMsg:  "ok",
		Data: &computeEngineAPIProtocol.GetPackageListASCByLastIdResponseData{
			PackageInfos: packageInfoList,
		},
	}
	slog.InfoContext(ctx, "GetPackageListASCByLastId", slog.Any("req", req), slog.Any("rsp", rsp))
	return rsp, nil
}

// RegisterPackage implements [DataManagementPlatform_ComputeEngineAPI.ComputeEngineAPIServer].
func (c *ComputeEngineService) RegisterPackage(ctx context.Context, req *computeEngineAPIProtocol.RegisterPackageRequest) (rsp *computeEngineAPIProtocol.RegisterPackageResponse, err error) {
	slog.InfoContext(ctx, "RegisterPackage", slog.Any("package_info", req))
	err = req.Validate()
	if err != nil {
		errMsg := "invalid request parameters: " + err.Error()
		slog.ErrorContext(ctx, errMsg, slog.Any("request", req.String()))
		return nil, status.Errorf(codes.InvalidArgument, "invalid request parameters")
	}

	rsp = &computeEngineAPIProtocol.RegisterPackageResponse{}
	return rsp, nil
}

func NewComputeEngineService(packageLogic *logic.PackageLogic) *ComputeEngineService {
	computeEngineServiceOnce.Do(func() {
		computeEngineServiceInstance = &ComputeEngineService{
			packageLogic: packageLogic,
		}
	})
	return computeEngineServiceInstance
}

func convertPackageEntityToBasicPackageInfo(ctx context.Context, packageEntity *entity.PackageEntity) *computeEngineAPIProtocol.BasicPackageInfo {
	if packageEntity == nil {
		return nil
	}

	packageInfo := &computeEngineAPIProtocol.BasicPackageInfo{
		Id:           packageEntity.ID,
		Name:         packageEntity.Name,
		Description:  packageEntity.Description,
		Platform:     packageEntity.Platform,
		Url:          packageEntity.Url,
		RegisterTime: packageEntity.RegisterTime.UnixMilli(),
	}
	return packageInfo
}
