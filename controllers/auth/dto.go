package auth

import "go-gin-api/models"

type RegisterDTO struct {
	Name     string `json:"name" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

type RoleDTO struct {
	Name        string   `json:"name"`
	Permissions []string `json:"permissions"`
}

type MeResponseDTO struct {
	Name  string    `json:"name"`
	Email string    `json:"email"`
	Roles []RoleDTO `json:"roles"`
}

func ToUserDTO(user models.User) MeResponseDTO {
	roles := make([]RoleDTO, len(user.Roles))

	for i, r := range user.Roles {
		permissions := make([]string, len(r.Permissions))
		for j, p := range r.Permissions {
			permissions[j] = p.Name
		}
		roles[i] = RoleDTO{Name: r.Name, Permissions: permissions}
	}

	return MeResponseDTO{
		Name:  user.Name,
		Email: user.Email,
		Roles: roles,
	}
}
