package admissionbridge

import (
	"context"
	"errors"
	"strings"

	"github.com/songconmaisaix31-design/Astrocyte/internal/apierrors"
	attention "github.com/songconmaisaix31-design/Astrocyte/internal/attention/app"
	workspace "github.com/songconmaisaix31-design/Astrocyte/internal/workspace/app"
	"github.com/songconmaisaix31-design/Astrocyte/internal/workspace/domain"
)

type humanProjectAccess struct {
	projects workspace.LocalProjects
}

var _ attention.AdmissionProjectAccess = (*humanProjectAccess)(nil)

func NewHumanProjectAccess(projects workspace.LocalProjects) (attention.AdmissionProjectAccess, error) {
	if projects == nil {
		return nil, errors.New("human admission requires the actual local project service")
	}
	return &humanProjectAccess{projects: projects}, nil
}

// 准入服务在同一事务核实固定资料；此端口核实当前人工主体和注册项目。
func (a *humanProjectAccess) AuthorizeAdmissionProject(ctx context.Context, principal attention.Principal, projectID string, _ []attention.SourceRef) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	denied := &apierrors.ServiceError{Code: apierrors.ScopeDenied, Message: "Human admission requires the current registered local project identity", RequiredAction: "select_registered_project"}
	if principal.Kind != "human" || strings.TrimSpace(principal.ID) == "" || projectID == "" || strings.TrimSpace(projectID) != projectID {
		return denied
	}
	projects, err := a.projects.ListProjects(ctx, domain.Caller{Kind: principal.Kind, ID: principal.ID})
	if err != nil {
		return err
	}
	for _, project := range projects {
		if project.ID == projectID {
			return nil
		}
	}
	return denied
}
