package db

import (
	"database/sql"
	"errors"
	"fmt"
	sq "github.com/Masterminds/squirrel"
	"go-service-template/domain"
	"go-service-template/monitor"
	"go.opentelemetry.io/otel/codes"
)

type LocationsRepository struct {
	queryBuilder sq.StatementBuilderType
	*TxDBContext
}

func (dal *LocationsRepository) CreateLocation(ctx monitor.ApplicationContext, location domain.Location) error {
	ctx, span := ctx.StartSpan("LocationsRepository.CreateLocation")
	defer span.End()

	_, err := dal.Exec(
		ctx,
		InsertLocation,
		location.ID,
		location.Name,
		location.LocationType.ID,
		location.Supplier.ID,
		location.Active,
	)
	if err != nil {
		return err
	}

	_, err = dal.Exec(
		ctx,
		InsertLocationInformation,
		location.Information.ID,
		location.ID,
		location.Information.Address,
		location.Information.City,
		location.Information.State,
		location.Information.Zipcode,
		location.Information.ContactInformation.ContactPerson,
		location.Information.ContactInformation.PhoneNumber,
		location.Information.ContactInformation.Email,
		location.Information.Latitude,
		location.Information.Longitude,
	)
	if err != nil {
		return err
	}

	return nil
}

func (dal *LocationsRepository) UpdateLocation(ctx monitor.ApplicationContext, location domain.Location) error {
	ctx, span := ctx.StartSpan("LocationsRepository.UpdateLocation")
	defer span.End()

	_, err := dal.Exec(
		ctx,
		UpdateLocation,
		location.Name,
		location.LocationType.ID,
		location.Supplier.ID,
		location.Active,
		location.ID,
	)
	if err != nil {
		return err
	}

	_, err = dal.Exec(
		ctx,
		UpdateLocationInformation,
		location.Information.Address,
		location.Information.City,
		location.Information.State,
		location.Information.Zipcode,
		location.Information.ContactInformation.ContactPerson,
		location.Information.ContactInformation.PhoneNumber,
		location.Information.ContactInformation.Email,
		location.Information.Latitude,
		location.Information.Longitude,
		location.Information.ID,
	)
	if err != nil {
		return err
	}

	return nil
}

func (dal *LocationsRepository) CreateSubLocation(ctx monitor.ApplicationContext, subLocation domain.SubLocation) error {
	ctx, span := ctx.StartSpan("LocationsRepository.CreateSubLocation")
	defer span.End()

	_, err := dal.Exec(
		ctx,
		InsertSubLocation,
		subLocation.ID,
		subLocation.LocationID,
		subLocation.SubLocationType.ID,
		subLocation.Name,
		subLocation.Active,
	)

	return err
}

func (dal *LocationsRepository) GetLocationByID(ctx monitor.ApplicationContext, id string) (*domain.Location, error) {
	ctx, span := ctx.StartSpan("LocationsRepository.GetLocationByID")
	defer span.End()

	return dal.parseLocationFromRow(dal.getDBReader().QueryRowContext(ctx, GetLocationByID, id))
}

func (dal *LocationsRepository) CheckLocationNameExistence(ctx monitor.ApplicationContext, name string) (bool, error) {
	ctx, span := ctx.StartSpan("LocationsRepository.CheckLocationNameExistence")
	defer span.End()

	var locationID string

	if err := dal.getDBReader().QueryRowContext(ctx, CheckLocationNameExistence, name).Scan(&locationID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}

		return false, err
	}

	return true, nil
}

// nolint
func (dal *LocationsRepository) GetPaginatedLocations(ctx monitor.ApplicationContext, filters domain.LocationsFilters) (domain.CursorPage[domain.Location], error) {
	ctx, span := ctx.StartSpan("LocationsRepository.GetPaginatedLocations")
	defer span.End()

	var result domain.CursorPage[domain.Location]

	// Build base query
	baseSelectQuery := dal.queryBuilder.Select(
		"l.id",
		"l.name",
		"l.active",
		"s.id",
		"s.name",
		"lt.id",
		"lt.type",
		"li.id",
		"li.address",
		"li.city",
		"li.state",
		"li.zipcode",
		"li.contact_person",
		"li.phone_number",
		"li.email",
		"li.latitude",
		"li.longitude",
	).From("location.locations l").InnerJoin(
		"location.location_information li on l.id = li.location_id",
	).InnerJoin(
		"location.location_types lt on l.location_type_id = lt.id",
	).InnerJoin(
		"location.suppliers s on s.id = l.supplier_id",
	)

	// Add filters
	if filters.Name != nil {
		filterClause := "l.name ILIKE CONCAT ('%',?::text,'%')"
		baseSelectQuery = baseSelectQuery.Where(filterClause, *filters.Name)
	}

	// Pagination filters
	if filters.CursorPaginationFilters.Cursor == "" {
		baseSelectQuery = baseSelectQuery.OrderBy("l.name ASC").Limit(uint64(filters.Limit) + 1)
	} else {
		var paginationClause, orderClause string

		if filters.CursorPaginationFilters.Direction == domain.NextPage {
			paginationClause = "l.name > ?"
			orderClause = "l.name ASC"
		} else {
			paginationClause = "l.name < ?"
			orderClause = "l.name DESC"
		}
		baseSelectQuery = baseSelectQuery.Where(paginationClause, filters.Cursor).OrderBy(orderClause).Limit(uint64(filters.Limit) + 1)
	}

	selectQueryStr, args, err := baseSelectQuery.ToSql()
	if err != nil {
		return result, fmt.Errorf("error when building GetPaginatedLocations query: %w", err)
	}

	rows, err := dal.getDBReader().QueryContext(ctx, selectQueryStr, args...)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		return result, err
	}
	defer rows.Close()

	locations := make([]domain.Location, 0)
	for rows.Next() {
		var location domain.Location
		if err := rows.Scan(
			&location.ID,
			&location.Name,
			&location.Active,
			&location.Supplier.ID,
			&location.Supplier.Name,
			&location.LocationType.ID,
			&location.LocationType.Type,
			&location.Information.ID,
			&location.Information.Address,
			&location.Information.City,
			&location.Information.State,
			&location.Information.Zipcode,
			&location.Information.ContactInformation.ContactPerson,
			&location.Information.ContactInformation.PhoneNumber,
			&location.Information.ContactInformation.Email,
			&location.Information.Latitude,
			&location.Information.Longitude,
		); err != nil {
			span.SetStatus(codes.Error, err.Error())
			return result, err
		}
		locations = append(locations, location)
	}

	result = domain.BuildCursorPage(locations, filters.CursorPaginationFilters)

	return result, nil
}

// nolint
func (dal *LocationsRepository) parseLocationFromRow(row *sql.Row) (*domain.Location, error) {
	var location domain.Location

	if err := row.Scan(
		&location.ID,
		&location.Name,
		&location.Active,
		&location.Supplier.ID,
		&location.Supplier.Name,
		&location.LocationType.ID,
		&location.LocationType.Type,
		&location.Information.ID,
		&location.Information.Address,
		&location.Information.City,
		&location.Information.State,
		&location.Information.Zipcode,
		&location.Information.ContactInformation.ContactPerson,
		&location.Information.ContactInformation.PhoneNumber,
		&location.Information.ContactInformation.Email,
		&location.Information.Latitude,
		&location.Information.Longitude,
	); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}

		return nil, err
	}

	return &location, nil
}
