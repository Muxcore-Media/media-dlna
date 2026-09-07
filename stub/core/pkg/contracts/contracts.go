package contracts

import "context"

// Module is the MuxCore sidecar lifecycle contract.
type Module interface {
	Info() ModuleInfo
	Init(ctx context.Context) error
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
	Health(ctx context.Context) error
}

// ContractDeclaration documents an external interface dependency.
type ContractDeclaration struct {
	Repo        string
	Interface   string
	Version     string
	Description string
}

// ModuleInfo is advertised to the mesh during registration.
type ModuleInfo struct {
	ID             string
	Name           string
	Version        string
	Description    string
	Author         string
	Roles          []string
	Capabilities   []string
	Contracts      []ContractDeclaration
	MinCoreVersion string
	HTTPAddr       string
}
