package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sync"

	"github.com/KitHub/DataManagementPlatform_ComputeEngineAPI/entity"
	"github.com/KitHub/DataManagementPlatform_ComputeEngineAPI/logic"
	computeEngineAPIProtocol "github.com/KitHub/protocols/DataManagementPlatform_ComputeEngineAPI"
	"github.com/KitHub/protocols/compute_operation"
	"google.golang.org/grpc/codes"
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

func (c *ComputeEngineService) ComputePackagesCombo(ctx context.Context, req *computeEngineAPIProtocol.ComputePackageComboRequest) (rsp *computeEngineAPIProtocol.ComputePackageComboResponse, err error) {
	slog.InfoContext(ctx, "computeCombo", slog.Any("req", req))

	err = c.validateRequest(ctx, req)
	if err != nil {
		slog.ErrorContext(ctx, "ComputePackagesCombo request validate failed", slog.Any("req", req), slog.Any("error", err))
		rsp = createPBRspWithPBMessageType[computeEngineAPIProtocol.ComputePackageComboResponse](ctx, codes.InvalidArgument, nil)
		return rsp, nil
	}

	root, err := convertComputeComboReqToSetOperationsRoot(ctx, req.GetCombo())
	if err != nil {
		slog.ErrorContext(ctx, "convertComputeComboReqToSetOperationsRoot failed", slog.Any("req", req), slog.Any("error", err))
		rsp = createPBRspWithPBMessageType[computeEngineAPIProtocol.ComputePackageComboResponse](ctx, codes.Internal, nil)
		return rsp, nil
	}

	resultOriginId, err := c.computeLogic.ComputeCombo(ctx, root)
	if err != nil {
		slog.InfoContext(ctx, "compute packages combo failed", slog.Any("setOperationNodes", root), slog.Any("error", err))
		rsp = createPBRspWithPBMessageType[computeEngineAPIProtocol.ComputePackageComboResponse](ctx, codes.Internal, nil)
		return rsp, nil
	}

	rsp = createPBRspWithPBMessageType[computeEngineAPIProtocol.ComputePackageComboResponse](ctx, codes.OK, &computeEngineAPIProtocol.ComputePackageComboResponse{
		ErrCode: 0,
		ErrMsg:  "",
		Data: &computeEngineAPIProtocol.ComputePackageComboResponseData{
			PackageOriginId: resultOriginId,
		},
	})

	slog.InfoContext(ctx, "computeCombo done", slog.Any("req", req), slog.String("resultOriginId", resultOriginId))
	return rsp, nil
}

func (c *ComputeEngineService) ReloadPackage(ctx context.Context, req *computeEngineAPIProtocol.ReloadPackageRequest) (rsp *computeEngineAPIProtocol.ReloadPackageResponse, err error) {
	// improve performance, reload in other routine
	slog.InfoContext(ctx, "ReloadPackage", slog.Any("req", req))

	err = req.Validate()
	if err != nil {
		slog.ErrorContext(ctx, "ReloadPackage request validate failed", slog.Any("req", req), slog.Any("error", err))
		rsp = createPBRspWithPBMessageType[computeEngineAPIProtocol.ReloadPackageResponse](ctx, codes.InvalidArgument, nil)
		return rsp, nil
	}

	packageEntity, err := c.packageLogic.QueryPackageByOriginId(ctx, req.GetOriginId())
	if err != nil {
		slog.ErrorContext(ctx, "ReloadPackage failed, querying package failed", slog.String("packageOriginId", req.GetOriginId()), slog.Any("error", err))
		rsp = createPBRspWithPBMessageType[computeEngineAPIProtocol.ReloadPackageResponse](ctx, codes.Internal, nil)
		return rsp, nil
	}

	err = c.computeLogic.ReloadPackage(ctx, packageEntity, true)
	if err != nil {
		slog.ErrorContext(ctx, "ReloadPackage failed", slog.String("packageOriginId", req.GetOriginId()), slog.Any("error", err))
		rsp = createPBRspWithPBMessageType[computeEngineAPIProtocol.ReloadPackageResponse](ctx, codes.Internal, nil)
		return rsp, nil
	}

	rsp = createPBRspWithPBMessageType[computeEngineAPIProtocol.ReloadPackageResponse](ctx, codes.OK, nil)
	slog.InfoContext(ctx, "ReloadPackage done")
	return rsp, nil
}

func (c *ComputeEngineService) ReloadAllPackage(ctx context.Context, req *computeEngineAPIProtocol.ReloadAllPackagesRequest) (rsp *computeEngineAPIProtocol.ReloadAllPackagesResponse, err error) {
	// improve performance, reload in other routine
	slog.InfoContext(ctx, "ReloadAllPackage", slog.Any("req", req))
	err = c.computeLogic.ReloadAllPackages(ctx, true)
	if err != nil {
		slog.ErrorContext(ctx, "ReloadAllPackage failed", slog.Any("error", err))
		rsp = createPBRspWithPBMessageType[computeEngineAPIProtocol.ReloadAllPackagesResponse](ctx, codes.Internal, nil)
		return rsp, nil
	}

	rsp = createPBRspWithPBMessageType[computeEngineAPIProtocol.ReloadAllPackagesResponse](ctx, codes.OK, nil)
	slog.InfoContext(ctx, "ReloadAllPackage done")
	return rsp, nil
}

// UploadPackage implements [DataManagementPlatform_ComputeEngineAPI.ComputeEngineAPIServer].
func (c *ComputeEngineService) UploadPackage(ctx context.Context, req *computeEngineAPIProtocol.UploadPackageRequest) (rsp *computeEngineAPIProtocol.UploadPackageResponse, err error) {
	slog.InfoContext(ctx, "UploadPackage", slog.Any("package_info", req))
	err = req.Validate()
	if err != nil {
		errMsg := "invalid request parameters: " + err.Error()
		slog.ErrorContext(ctx, errMsg, slog.Any("request", req.String()))
		rsp = createPBRspWithPBMessageType[computeEngineAPIProtocol.UploadPackageResponse](ctx, codes.InvalidArgument, nil)
		return rsp, nil
	}

	packageEntity, err := c.packageLogic.QueryPackageByOriginId(ctx, req.GetOriginId())
	if err != nil {
		slog.ErrorContext(ctx, "queryPackageById failed", slog.Any("err", err))
		rsp = createPBRspWithPBMessageType[computeEngineAPIProtocol.UploadPackageResponse](ctx, codes.Internal, nil)
		return rsp, nil
	}

	if packageEntity != nil {
		errMsg := "package with the same origin_id existed"
		slog.ErrorContext(ctx, errMsg, slog.String("origin_id", req.GetOriginId()))
		rsp = createPBRspWithPBMessageType[computeEngineAPIProtocol.UploadPackageResponse](ctx, codes.Internal, nil)
		return rsp, nil
	}

	err = c.packageLogic.UploadPackageFromMemory(ctx, req.GetBucketName(), req.GetKeyName(), req.GetPackageFile())
	if err != nil {
		slog.ErrorContext(ctx, "upload package failed", slog.String("bucket_name", req.GetBucketName()), slog.String("key_name", req.GetKeyName()))
		rsp = createPBRspWithPBMessageType[computeEngineAPIProtocol.UploadPackageResponse](ctx, codes.Internal, nil)
		return rsp, nil
	}

	newPackageEntity, err := c.packageLogic.InsertPackage(ctx, req.GetOriginId(), req.GetDisplayName(), req.GetComment(), req.GetPlatform(), req.GetBucketName(), req.GetKeyName())
	if err != nil {
		slog.ErrorContext(ctx, "insert package failed", slog.Any("req", req))
		rsp = createPBRspWithPBMessageType[computeEngineAPIProtocol.UploadPackageResponse](ctx, codes.Internal, nil)
		return rsp, nil
	}

	rsp = createPBRspWithPBMessageType[computeEngineAPIProtocol.UploadPackageResponse](ctx, codes.OK, &computeEngineAPIProtocol.UploadPackageResponseData{
		PackageInfo: convertPackageEntityToBasicPackageInfo(ctx, newPackageEntity),
	})

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
		rsp = createPBRspWithPBMessageType[computeEngineAPIProtocol.GetPackageByIdResponse](ctx, codes.InvalidArgument, nil)
		return rsp, nil
	}

	packageEntity, err := c.packageLogic.QueryPackageById(ctx, req.GetPackageId())
	if err != nil {
		slog.ErrorContext(ctx, "queryPackageById failed", slog.Any("err", err))
		rsp = createPBRspWithPBMessageType[computeEngineAPIProtocol.GetPackageByIdResponse](ctx, codes.Internal, nil)
		return rsp, nil
	}

	rsp = createPBRspWithPBMessageType[computeEngineAPIProtocol.GetPackageByIdResponse](ctx, codes.OK, &computeEngineAPIProtocol.GetPackageByIdResponseData{PackageInfo: convertPackageEntityToBasicPackageInfo(ctx, packageEntity)})

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
		rsp = createPBRspWithPBMessageType[computeEngineAPIProtocol.GetPackageByOriginIdResponse](ctx, codes.InvalidArgument, nil)
		return rsp, nil
	}

	packageEntity, err := c.packageLogic.QueryPackageByOriginId(ctx, req.GetOriginId())
	if err != nil {
		slog.ErrorContext(ctx, "queryPackageById failed", slog.Any("err", err))
		rsp = createPBRspWithPBMessageType[computeEngineAPIProtocol.GetPackageByOriginIdResponse](ctx, codes.Internal, nil)
		return rsp, nil
	}

	rsp = createPBRspWithPBMessageType[computeEngineAPIProtocol.GetPackageByOriginIdResponse](ctx, codes.InvalidArgument, &computeEngineAPIProtocol.GetPackageByOriginIdResponseData{
		PackageInfo: convertPackageEntityToBasicPackageInfo(ctx, packageEntity),
	})

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
		rsp = createPBRspWithPBMessageType[computeEngineAPIProtocol.GetPackageListASCByLastIdResponse](ctx, codes.InvalidArgument, nil)
		return rsp, nil
	}

	packageEntityList, err := c.packageLogic.QueryPackageListASCById(ctx, req.GetLastId(), req.GetLimit())
	if err != nil {
		slog.ErrorContext(ctx, "QueryPackageListASCById failed", slog.Any("err", err))
		rsp = createPBRspWithPBMessageType[computeEngineAPIProtocol.GetPackageListASCByLastIdResponse](ctx, codes.Internal, nil)
		return rsp, nil
	}

	packageInfoList := make([]*computeEngineAPIProtocol.BasicPackageInfo, 0)
	for _, tmpPackageEntity := range packageEntityList {
		packageInfoList = append(packageInfoList, convertPackageEntityToBasicPackageInfo(ctx, tmpPackageEntity))
	}

	rsp = createPBRspWithPBMessageType[computeEngineAPIProtocol.GetPackageListASCByLastIdResponse](ctx, codes.InvalidArgument, &computeEngineAPIProtocol.GetPackageListASCByLastIdResponseData{
		PackageInfos: packageInfoList,
	})
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
		rsp = createPBRspWithPBMessageType[computeEngineAPIProtocol.RegisterPackageResponse](ctx, codes.InvalidArgument, nil)
		return rsp, nil
	}

	packageEntity := &entity.PackageEntity{
		OriginId:    req.GetOriginId(),
		DisplayName: req.GetDisplayName(),
		Comment:     req.GetComment(),
		Platform:    req.GetPlatform(),
		BucketName:  req.GetBucketName(),
		KeyName:     req.GetKeyName(),
	}

	packageEntity, err = c.packageLogic.InsertPackage(ctx, packageEntity.OriginId, packageEntity.DisplayName, packageEntity.Comment, packageEntity.Platform, packageEntity.BucketName, packageEntity.KeyName)
	if err != nil {
		slog.ErrorContext(ctx, "insert package info failed", slog.Any("packageEntity", packageEntity), slog.Any("error", err))
		rsp = createPBRspWithPBMessageType[computeEngineAPIProtocol.RegisterPackageResponse](ctx, codes.Internal, nil)
		return rsp, nil
	}

	rsp = createPBRspWithPBMessageType[computeEngineAPIProtocol.RegisterPackageResponse](ctx, codes.OK, &computeEngineAPIProtocol.RegisterPackageResponseData{
		PackageInfo: convertPackageEntityToBasicPackageInfo(ctx, packageEntity),
	})

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

func (c *ComputeEngineService) validateRequest(ctx context.Context, req *computeEngineAPIProtocol.ComputePackageComboRequest) error {
	if err := req.Validate(); err != nil {
		return err
	}
	if err := c.validateAstNode(ctx, req.GetCombo(), 8, 1); err != nil {
		return err
	}
	return nil
}

func (c *ComputeEngineService) validateAstNode(ctx context.Context, node *compute_operation.SetCompute_Node, maxDepth int32, currentDepth int32) error {
	if currentDepth > maxDepth {
		return errors.New("combo layers is too deep, max depth is " + fmt.Sprintf("%d", maxDepth))
	}
	switch node.GetNodeType() {
	case compute_operation.SetCompute_DATASET:
		if node.SetKey == "" {
			return errors.New("setKey cannot be empty for DataSet node")
		}
		if node.Operator != compute_operation.SetCompute_OPERATOR_UNSPECIFIED || len(node.Children) > 0 {
			return errors.New("DataSet node should not have operator or children")
		}
		if ok := c.computeLogic.ValidatePackage(ctx, node.SetKey); !ok {
			return fmt.Errorf("%s: %s", "package not found", node.SetKey)
		}
		return nil

	case compute_operation.SetCompute_OPERATOR:
		if node.Operator == compute_operation.SetCompute_OPERATOR_UNSPECIFIED {
			return errors.New("operator node must have a operator")
		}
		if len(node.Children) == 0 {
			return errors.New("operator node must have children")
		}
		switch node.Operator {
		case compute_operation.SetCompute_INTERSECT, compute_operation.SetCompute_UNION:
			if len(node.Children) < 2 {
				return fmt.Errorf("%s: union/intersect must have at least 2 children", "ErrChildCountIllegal")
			}
		case compute_operation.SetCompute_DIFF:
			if len(node.Children) != 2 {
				return fmt.Errorf("%s: diff must have exactly 2 children", "ErrChildCountIllegal")
			}
		}
		for _, child := range node.Children {
			if err := c.validateAstNode(ctx, child, maxDepth, currentDepth+1); err != nil {
				return err
			}
		}
		return nil

	default:
		return errors.New("not supported node type: " + node.GetNodeType().String())
	}
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

func convertComputeComboReqToSetOperationsRoot(ctx context.Context, combo *compute_operation.SetCompute_Node) (root *entity.SetOperationNode, err error) {
	comboStack := []*compute_operation.SetCompute_Node{combo}
	mapping := make(map[*compute_operation.SetCompute_Node]*entity.SetOperationNode)
	for len(comboStack) > 0 {
		currentComboNode := comboStack[len(comboStack)-1]
		comboStack = comboStack[:len(comboStack)-1]

		newOperationNode := &entity.SetOperationNode{
			Data:        currentComboNode.SetKey,
			SetOperator: entity.SetOperator(currentComboNode.Operator),
		}

		mapping[currentComboNode] = newOperationNode

		for _, child := range currentComboNode.Children {
			comboStack = append(comboStack, child)
		}
	}

	comboStack = []*compute_operation.SetCompute_Node{combo}
	for len(comboStack) > 0 {
		currentComboNode := comboStack[len(comboStack)-1]
		comboStack = comboStack[:len(comboStack)-1]

		newOperationNode := mapping[currentComboNode]
		newOperationNode.Children = make([]*entity.SetOperationNode, 0, len(currentComboNode.Children))
		for _, child := range currentComboNode.Children {
			newOperationNode.Children = append(newOperationNode.Children, mapping[child])
			comboStack = append(comboStack, child)
		}
	}

	root = mapping[combo]
	return root, nil
}
