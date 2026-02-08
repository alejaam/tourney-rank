// Package team provides use cases for team management.
package team

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/melisource/tourney-rank/internal/domain/player"
	"github.com/melisource/tourney-rank/internal/domain/team"
	"github.com/melisource/tourney-rank/internal/domain/tournament"
)

// Service handles team use cases.
type Service struct {
	teamRepo       team.Repository
	tournamentRepo tournament.Repository
	playerRepo     player.Repository
}

// NewService creates a new team service.
func NewService(teamRepo team.Repository, tournamentRepo tournament.Repository, playerRepo player.Repository) *Service {
	return &Service{
		teamRepo:       teamRepo,
		tournamentRepo: tournamentRepo,
		playerRepo:     playerRepo,
	}
}

// CreateTeamRequest represents the request to create a team.
type CreateTeamRequest struct {
	TournamentID uuid.UUID `json:"tournament_id"`
	Name         string    `json:"name"`
	Tag          string    `json:"tag,omitempty"`
	LogoURL      string    `json:"logo_url,omitempty"`
}

// TeamMemberInfo represents information about a team member.
type TeamMemberInfo struct {
	PlayerID    uuid.UUID `json:"player_id"`
	DisplayName string    `json:"display_name"`
	AvatarURL   string    `json:"avatar_url"`
	IsCaptain   bool      `json:"is_captain"`
}

// TeamWithMembers represents a team with full member information.
type TeamWithMembers struct {
	*team.Team
	Members []*TeamMemberInfo `json:"members"`
}

// JoinTeamRequest represents the request to join a team via invite code.
type JoinTeamRequest struct {
	InviteCode string `json:"invite_code"`
}

// RemoveMemberRequest represents the request to remove a member from a team.
type RemoveMemberRequest struct {
	PlayerID uuid.UUID `json:"player_id"`
}

// TransferCaptaincyRequest represents the request to transfer team captaincy.
type TransferCaptaincyRequest struct {
	NewCaptainID uuid.UUID `json:"new_captain_id"`
}

// UpdateTeamRequest represents the request to update a team.
type UpdateTeamRequest struct {
	Name    *string `json:"name,omitempty"`
	Tag     *string `json:"tag,omitempty"`
	LogoURL *string `json:"logo_url,omitempty"`
}

// CreateTeam creates a new team.
func (s *Service) CreateTeam(ctx context.Context, req CreateTeamRequest, captainID uuid.UUID) (*team.Team, error) {
	// Verify tournament exists and is open for registration
	t, err := s.tournamentRepo.GetByID(ctx, req.TournamentID)
	if err != nil {
		return nil, err
	}

	if t.Status != tournament.StatusOpen && !t.Rules.AllowLateRegistration {
		return nil, tournament.ErrRegistrationClosed
	}

	// Verify player exists
	_, err = s.playerRepo.GetByID(ctx, captainID.String())
	if err != nil {
		return nil, err
	}

	// Check if player already has a team in this tournament
	existingTeam, err := s.GetPlayerTeamInTournament(ctx, captainID, req.TournamentID)
	if err == nil && existingTeam != nil {
		return nil, team.ErrPlayerAlreadyInTeam
	}

	tm, err := team.NewTeam(req.TournamentID, captainID, req.Name)
	if err != nil {
		return nil, err
	}

	if req.Tag != "" {
		tm.SetTag(req.Tag)
	}
	if req.LogoURL != "" {
		tm.SetLogoURL(req.LogoURL)
	}

	if err := s.teamRepo.Create(ctx, tm); err != nil {
		return nil, err
	}

	return tm, nil
}

// GetTeam retrieves a team by ID.
func (s *Service) GetTeam(ctx context.Context, id uuid.UUID) (*team.Team, error) {
	return s.teamRepo.GetByID(ctx, id)
}

// GetTeamByInviteCode retrieves a team by its invite code.
func (s *Service) GetTeamByInviteCode(ctx context.Context, inviteCode string) (*team.Team, error) {
	return s.teamRepo.GetByInviteCode(ctx, inviteCode)
}

// GetTeamWithMembers retrieves a team with full member information.
func (s *Service) GetTeamWithMembers(ctx context.Context, id uuid.UUID) (*TeamWithMembers, error) {
	tm, err := s.teamRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	members := make([]*TeamMemberInfo, 0, len(tm.MemberIDs))
	for _, memberID := range tm.MemberIDs {
		p, err := s.playerRepo.GetByID(ctx, memberID.String())
		if err != nil {
			continue // Skip if player not found
		}

		members = append(members, &TeamMemberInfo{
			PlayerID:    p.ID,
			DisplayName: p.DisplayName,
			AvatarURL:   p.AvatarURL,
			IsCaptain:   tm.IsCaptain(p.ID),
		})
	}

	return &TeamWithMembers{
		Team:    tm,
		Members: members,
	}, nil
}

// JoinTeam allows a player to join a team via invite code.
func (s *Service) JoinTeam(ctx context.Context, req JoinTeamRequest, playerID uuid.UUID) (*team.Team, error) {
	// Get team by invite code
	tm, err := s.teamRepo.GetByInviteCode(ctx, req.InviteCode)
	if err != nil {
		return nil, team.ErrInvalidInviteCode
	}

	// Verify player exists
	_, err = s.playerRepo.GetByID(ctx, playerID.String())
	if err != nil {
		return nil, err
	}

	// Check if player already in team
	if tm.HasMember(playerID) {
		return nil, team.ErrPlayerAlreadyInTeam
	}

	// Check if player already has a team in this tournament
	existingTeam, err := s.GetPlayerTeamInTournament(ctx, playerID, tm.TournamentID)
	if err == nil && existingTeam != nil {
		return nil, team.ErrPlayerAlreadyInTeam
	}

	// Verify tournament allows registration
	t, err := s.tournamentRepo.GetByID(ctx, tm.TournamentID)
	if err != nil {
		return nil, err
	}

	if t.Status != tournament.StatusOpen && !t.Rules.AllowLateRegistration {
		return nil, tournament.ErrRegistrationClosed
	}

	// Check team size limit
	if tm.MemberCount() >= int(t.TeamSize) {
		return nil, team.ErrTeamFull
	}

	// Add member to team
	if err := tm.AddMember(playerID); err != nil {
		return nil, err
	}

	// Update team in repository
	if err := s.teamRepo.Update(ctx, tm); err != nil {
		return nil, err
	}

	return tm, nil
}

// RemoveMember removes a member from a team.
func (s *Service) RemoveMember(ctx context.Context, teamID uuid.UUID, req RemoveMemberRequest, requestorID uuid.UUID) (*team.Team, error) {
	tm, err := s.teamRepo.GetByID(ctx, teamID)
	if err != nil {
		return nil, err
	}

	// Only captain can remove members
	if !tm.IsCaptain(requestorID) {
		return nil, team.ErrNotCaptain
	}

	if err := tm.RemoveMember(req.PlayerID); err != nil {
		return nil, err
	}

	if err := s.teamRepo.Update(ctx, tm); err != nil {
		return nil, err
	}

	return tm, nil
}

// LeaveTeam allows a player to leave a team.
func (s *Service) LeaveTeam(ctx context.Context, teamID, playerID uuid.UUID) error {
	tm, err := s.teamRepo.GetByID(ctx, teamID)
	if err != nil {
		return err
	}

	// Captain cannot leave unless transferring captaincy first
	if tm.IsCaptain(playerID) {
		return team.ErrCannotRemoveCaptain
	}

	if err := tm.RemoveMember(playerID); err != nil {
		return err
	}

	return s.teamRepo.Update(ctx, tm)
}

// TransferCaptaincy transfers team captaincy to another member.
func (s *Service) TransferCaptaincy(ctx context.Context, teamID uuid.UUID, req TransferCaptaincyRequest, requestorID uuid.UUID) (*team.Team, error) {
	tm, err := s.teamRepo.GetByID(ctx, teamID)
	if err != nil {
		return nil, err
	}

	// Only current captain can transfer captaincy
	if !tm.IsCaptain(requestorID) {
		return nil, team.ErrNotCaptain
	}

	if err := tm.TransferCaptaincy(req.NewCaptainID); err != nil {
		return nil, err
	}

	if err := s.teamRepo.Update(ctx, tm); err != nil {
		return nil, err
	}

	return tm, nil
}

// ListTeamsByTournament lists all teams in a tournament.
func (s *Service) ListTeamsByTournament(ctx context.Context, tournamentID uuid.UUID) ([]*team.Team, error) {
	return s.teamRepo.GetByTournamentID(ctx, tournamentID)
}

// GetPlayerTeamInTournament retrieves the team a player belongs to in a specific tournament.
func (s *Service) GetPlayerTeamInTournament(ctx context.Context, playerID, tournamentID uuid.UUID) (*team.Team, error) {
	teams, err := s.teamRepo.GetByTournamentID(ctx, tournamentID)
	if err != nil {
		return nil, err
	}

	for _, tm := range teams {
		if tm.HasMember(playerID) {
			return tm, nil
		}
	}

	return nil, team.ErrNotFound
}

// GetPlayerTeams retrieves all teams a player belongs to.
func (s *Service) GetPlayerTeams(ctx context.Context, playerID uuid.UUID) ([]*team.Team, error) {
	return s.teamRepo.GetByPlayerID(ctx, playerID)
}

// UpdateTeam updates a team's information.
func (s *Service) UpdateTeam(ctx context.Context, teamID uuid.UUID, req UpdateTeamRequest, requestorID uuid.UUID) (*team.Team, error) {
	tm, err := s.teamRepo.GetByID(ctx, teamID)
	if err != nil {
		return nil, err
	}

	// Only captain can update team
	if !tm.IsCaptain(requestorID) {
		return nil, team.ErrNotCaptain
	}

	if req.Name != nil {
		tm.Name = *req.Name
	}
	if req.Tag != nil {
		tm.SetTag(*req.Tag)
	}
	if req.LogoURL != nil {
		tm.SetLogoURL(*req.LogoURL)
	}

	tm.UpdatedAt = time.Now().UTC()

	if err := s.teamRepo.Update(ctx, tm); err != nil {
		return nil, err
	}

	return tm, nil
}

// DisbandTeam disbands a team.
func (s *Service) DisbandTeam(ctx context.Context, teamID, requestorID uuid.UUID) error {
	tm, err := s.teamRepo.GetByID(ctx, teamID)
	if err != nil {
		return err
	}

	// Only captain can disband team
	if !tm.IsCaptain(requestorID) {
		return team.ErrNotCaptain
	}

	if err := tm.UpdateStatus(team.StatusDisbanded); err != nil {
		return err
	}

	return s.teamRepo.Update(ctx, tm)
}
