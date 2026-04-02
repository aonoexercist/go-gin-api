package user

import "go-gin-api/models"

type RoleDTO struct {
	Name        string   `json:"name"`
	Permissions []string `json:"permissions"`
}

type UserResponseDTO struct {
	ID    uint      `json:"id"`
	Name  string    `json:"name"`
	Email string    `json:"email"`
	Roles []RoleDTO `json:"roles"`
}

func ToUserDTO(user models.User) UserResponseDTO {
	roles := make([]RoleDTO, len(user.Roles))

	for i, r := range user.Roles {
		permissions := make([]string, len(r.Permissions))
		for j, p := range r.Permissions {
			permissions[j] = p.Name
		}
		roles[i] = RoleDTO{Name: r.Name, Permissions: permissions}
	}

	return UserResponseDTO{
		ID:    user.ID,
		Name:  user.Name,
		Email: user.Email,
		Roles: roles,
	}
}
