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
	computeLogic *logic.ComputeLogic
}

func (c *ComputeEngineService) ReloadPackage(ctx context.Context, req *computeEngineAPIProtocol.ReloadPackageRequest) (rsp *computeEngineAPIProtocol.ReloadPackageResponse, err error) {
	// improve performance, reload in other routine
	slog.InfoContext(ctx, "ReloadPackage", slog.Any("req", req))

	packageEntity, err := c.packageLogic.QueryPackageByOriginId(ctx, req.GetOriginId())
	if err != nil {
		slog.ErrorContext(ctx, "ReloadPackage failed, querying package failed", slog.String("packageOriginId", req.GetOriginId()), slog.Any("error", err))
		return &computeEngineAPIProtocol.ReloadPackageResponse{
			ErrCode: 1,
			ErrMsg:  "server error",
		}, err
	}

	err = c.computeLogic.ReloadPackage(ctx, packageEntity, true)
	if err != nil {
		slog.ErrorContext(ctx, "ReloadPackage failed", slog.String("packageOriginId", req.GetOriginId()), slog.Any("error", err))
		return &computeEngineAPIProtocol.ReloadPackageResponse{
			ErrCode: 1,
			ErrMsg:  "server error",
		}, err
	}

	slog.InfoContext(ctx, "ReloadPackage done")
	return &computeEngineAPIProtocol.ReloadPackageResponse{
		ErrCode: 0,
		ErrMsg:  "",
	}, nil
}

func (c *ComputeEngineService) ReloadAllPackage(ctx context.Context, req *computeEngineAPIProtocol.ReloadAllPackagesRequest) (rsp *computeEngineAPIProtocol.ReloadAllPackagesResponse, err error) {
	// improve performance, reload in other routine
	slog.InfoContext(ctx, "ReloadAllPackage", slog.Any("req", req))
	err = c.computeLogic.ReloadAllPackages(ctx, true)
	if err != nil {
		slog.ErrorContext(ctx, "ReloadAllPackage failed", slog.Any("error", err))
		return &computeEngineAPIProtocol.ReloadAllPackagesResponse{
			ErrCode: 1,
			ErrMsg:  "server error",
		}, err
	}
	slog.InfoContext(ctx, "ReloadAllPackage done")
	return &computeEngineAPIProtocol.ReloadAllPackagesResponse{
		ErrCode: 0,
		ErrMsg:  "",
	}, nil
}

// UploadPackage implements [DataManagementPlatform_ComputeEngineAPI.ComputeEngineAPIServer].
func (c *ComputeEngineService) UploadPackage(ctx context.Context, req *computeEngineAPIProtocol.UploadPackageRequest) (rsp *computeEngineAPIProtocol.UploadPackageResponse, err error) {
	slog.InfoContext(ctx, "UploadPackage", slog.Any("package_info", req))
	err = req.Validate()
	if err != nil {
		errMsg := "invalid request parameters: " + err.Error()
		slog.ErrorContext(ctx, errMsg, slog.Any("request", req.String()))
		return nil, status.Errorf(codes.InvalidArgument, "invalid request parameters")
	}

	packageEntity, err := c.packageLogic.QueryPackageByOriginId(ctx, req.GetOriginId())
	if err != nil {
		slog.ErrorContext(ctx, "queryPackageById failed", slog.Any("err", err))
		return nil, status.Error(codes.Internal, "server error")
	}

	if packageEntity != nil {
		errMsg := "package with the same origin_id existed"
		slog.ErrorContext(ctx, errMsg, slog.String("origin_id", req.GetOriginId()))
		return nil, status.Error(codes.AlreadyExists, errMsg)
	}

	err = c.packageLogic.UploadPackage(ctx, req.GetBucketName(), req.GetKeyName(), req.GetPackageFile())
	if err != nil {
		slog.ErrorContext(ctx, "upload package failed", slog.String("bucket_name", req.GetBucketName()), slog.String("key_name", req.GetKeyName()))
		return nil, status.Error(codes.Internal, "server error")
	}

	newPackageEntity, err := c.packageLogic.InsertPackage(ctx, req.GetOriginId(), req.GetComment(), req.GetPlatform(), req.GetBucketName(), req.GetKeyName())
	if err != nil {
		slog.ErrorContext(ctx, "insert package failed", slog.Any("req", req))
		return nil, status.Error(codes.Internal, "server error")
	}

	rsp = &computeEngineAPIProtocol.UploadPackageResponse{
		ErrCode: 0,
		ErrMsg:  "ok",
		Data: &computeEngineAPIProtocol.UploadPackageResponseData{
			PackageInfo: convertPackageEntityToBasicPackageInfo(ctx, newPackageEntity),
		},
	}

	slog.InfoContext(ctx, "GetPackageById", slog.Any("req", req), slog.Any("rsp", rsp))
	return rsp, nil
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

// GetPackageByOriginId implements [DataManagementPlatform_ComputeEngineAPI.ComputeEngineAPIServer].
func (c *ComputeEngineService) GetPackageByOriginId(ctx context.Context, req *computeEngineAPIProtocol.GetPackageByOriginIdRequest) (rsp *computeEngineAPIProtocol.GetPackageByOriginIdResponse, err error) {
	slog.InfoContext(ctx, "GetPackageByName", slog.Any("originId", req.GetOriginId()))
	err = req.Validate()
	if err != nil {
		errMsg := "invalid request parameters: " + err.Error()
		slog.ErrorContext(ctx, errMsg, slog.Any("request", req.String()))
		return nil, status.Errorf(codes.InvalidArgument, "invalid request parameters")
	}

	packageEntity, err := c.packageLogic.QueryPackageByOriginId(ctx, req.GetOriginId())
	if err != nil {
		slog.ErrorContext(ctx, "queryPackageById failed", slog.Any("err", err))
		return nil, status.Errorf(codes.Internal, "server error")
	}

	rsp = &computeEngineAPIProtocol.GetPackageByOriginIdResponse{
		ErrCode: 0,
		ErrMsg:  "ok",
		Data: &computeEngineAPIProtocol.GetPackageByOriginIdResponseData{
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

	packageEntity := &entity.PackageEntity{
		OriginId:   req.GetOriginId(),
		Comment:    req.GetComment(),
		Platform:   req.GetPlatform(),
		BucketName: req.GetBucketName(),
		KeyName:    req.GetKeyName(),
	}

	packageEntity, err = c.packageLogic.InsertPackage(ctx, packageEntity.OriginId, packageEntity.Comment, packageEntity.Platform, packageEntity.BucketName, packageEntity.KeyName)
	if err != nil {
		slog.ErrorContext(ctx, "insert package info failed", slog.Any("packageEntity", packageEntity), slog.Any("error", err))
		return nil, status.Errorf(codes.Internal, "server error")
	}

	rsp = &computeEngineAPIProtocol.RegisterPackageResponse{
		ErrCode: 0,
		ErrMsg:  "ok",
		Data: &computeEngineAPIProtocol.RegisterPackageResponseData{
			PackageInfo: convertPackageEntityToBasicPackageInfo(ctx, packageEntity),
		},
	}

	slog.InfoContext(ctx, "RegisterPackage", slog.Any("req", req), slog.Any("rsp", rsp))
	return rsp, nil
}

func NewComputeEngineService(ctx context.Context, packageLogic *logic.PackageLogic, computeLogic *logic.ComputeLogic) *ComputeEngineService {
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
		OriginId:     packageEntity.OriginId,
		Comment:      packageEntity.Comment,
		Platform:     packageEntity.Platform,
		BucketName:   packageEntity.BucketName,
		KeyName:      packageEntity.KeyName,
		RegisterTime: packageEntity.RegisterTime.UnixMilli(),
	}
	return packageInfo
}
