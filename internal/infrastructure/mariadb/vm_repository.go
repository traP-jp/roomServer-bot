package mariadb

import (
	"context"
	"database/sql"

	"github.com/trap-jp/roomserver-bot/internal/domain"
	"github.com/trap-jp/roomserver-bot/internal/infrastructure/mariadb/db"
)

type vmRepository struct {
	q  *db.Queries
	db *sql.DB
}

func NewVMRepository(conn *sql.DB) domain.VMRepository {
	return &vmRepository{
		q:  db.New(conn),
		db: conn,
	}
}

func (r *vmRepository) SaveInstance(ctx context.Context, inst *domain.Instance) error {
	return r.q.CreateInstance(ctx, db.CreateInstanceParams{
		Vmid:         inst.Vmid,
		UserID:       inst.UserID,
		TemplateVmid: inst.TemplateVmid,
		IpAddress:    sql.NullString{String: inst.IpAddress, Valid: inst.IpAddress != ""},
	})
}

func (r *vmRepository) FindInstancesByUserID(ctx context.Context, userID string) ([]domain.Instance, error) {
	instancesDB, err := r.q.GetInstanceByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	instances := make([]domain.Instance, 0, len(instancesDB))
	for _, i := range instancesDB {
		instances = append(instances, domain.Instance{
			Vmid:         i.Vmid,
			UserID:       i.UserID,
			TemplateVmid: i.TemplateVmid,
			IpAddress:    i.IpAddress.String,
		})
	}
	return instances, nil
}

func (r *vmRepository) GetAllTemplates(ctx context.Context) ([]domain.VmTemplate, error) {
	templatesDB, err := r.q.ListVMTemplates(ctx)
	if err != nil {
		return nil, err
	}
	templates := make([]domain.VmTemplate, 0, len(templatesDB))
	for _, t := range templatesDB {
		templates = append(templates, domain.VmTemplate{
			Vmid: t.Vmid,
			Name: t.Name,
		})
	}
	return templates, nil
}
