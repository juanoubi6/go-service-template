package db

const (
	InsertLocationInformation = `INSERT INTO location.location_information (
                                           id,
                                           location_id,
                                           address,
                                           city,
                                           state,
                                           zipcode,
                                           contact_person,
                                           phone_number,
                                           email,
                                           latitude,
                                           longitude
								) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11);`

	InsertLocation = `INSERT INTO location.locations (
                                id,
                                name,
                                location_type_id,
                                supplier_id,
                                active
							) VALUES ($1,$2,$3,$4,$5);`

	InsertSubLocation = `INSERT INTO location.sub_locations (
									id,
									location_id,
									sub_location_type_id,
									name,
                                    active
								) VALUES ($1,$2,$3,$4,$5);`

	UpdateLocation = `UPDATE location.locations SET
								name = $1,
								location_type_id = $2,
								supplier_id = $3,
								active = $4,
								updated_at= CURRENT_TIMESTAMP
							WHERE id = $5;`

	UpdateLocationInformation = `UPDATE location.location_information SET
								address = $1,
								city = $2,
								state = $3,
								zipcode = $4,
								contact_person = $5,
								phone_number = $6,
								email = $7,
								latitude = $8,
								longitude = $9
							WHERE id = $10;`

	GetLocationByID = `SELECT
							l.id,
							l.name,
							l.active,
							s.id,
							s.name,
							lt.id,
							lt.type,
							li.id,
							li.address,
							li.city,
							li.state,
							li.zipcode,
							li.contact_person,
							li.phone_number,
							li.email,
							li.latitude,
							li.longitude
						FROM location.locations l
						JOIN location.location_information li on l.id = li.location_id
						JOIN location.location_types lt on l.location_type_id = lt.id
						JOIN location.suppliers s on s.id = l.supplier_id
						WHERE l.id = $1
						LIMIT 1 FOR UPDATE`

	CheckLocationNameExistence = `SELECT id FROM location.locations WHERE LOWER(name) = LOWER($1)`
)
