package bakery

// Query List to Prepare
const (
	getUserByEmail  = "GetUserByEmail"
	qGetUserByEmail = `SELECT id, nama, email, password_hash, aktif FROM users WHERE email = ?`

	listUsers  = "ListUsers"
	qListUsers = `SELECT id, nama, email, dibuat_pada FROM users WHERE aktif = TRUE ORDER BY nama, email`
)
