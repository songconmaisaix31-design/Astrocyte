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
	projects   workspace.LocalProjects
	references attention.AttentionProjectReferences
}

var _ attention.AdmissionProjectAccess = (*humanProjectAccess)(nil)

func NewHumanProjectAccess(projects workspace.LocalProjects, references attention.AttentionProjectReferences) (attention.AdmissionProjectAccess, error) {
	if projects == nil || references == nil {
		return nil, errors.New("human admission requires actual project and fixed reference services")
	}
	return &humanProjectAccess{projects: projects, references: references}, nil
}

// 当前人工项目和每项固定资料均通过原生产端口核实。
func (a *humanProjectAccess) AuthorizeAdmissionProject(ctx context.Context, principal attention.Principal, projectID string, refs []attention.SourceRef) error {
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
			for _, ref := range refs {
				if _, err := a.references.ReadProjectReference(ctx, principal, project.SpaceID, ref, false); err != nil {
					return err
				}
			}
			return nil
		}
	}
	return denied
}
