/* HUMAN MODE - Truth Protocol Active
   System: Senior CTO-scientist reasoning mode engaged
   Reward = TRUTH_PROVEN via tests. Claims = PROVISIONAL until verified.
   This file protected by HUMAN-TRUTH protocol - see truth_protocol/README.md
*/
package service

import "ares_api/internal/models"

type UserService interface {
	Signup(username, email, password string) error
	Login(username, password string)(id uint,accessToken string, refreshToken string, err error)
	Refresh(refreshToken string) (string, error)
	GetUserByID(id uint) (*models.User, error)
}
