package round

import (
	"context"
	"errors"
	"fmt"
	"time"

	rounddomain "github.com/alejaam/tourney-rank/internal/domain/round"
	tournamentdomain "github.com/alejaam/tourney-rank/internal/domain/tournament"
	"github.com/google/uuid"
)

// Service handles round business logic.
type Service struct {
	roundRepo       rounddomain.Repository
	tournamentRepo  tournamentdomain.Repository
}

// NewService creates a new round service.
func NewService(roundRepo rounddomain.Repository, tournamentRepo tournamentdomain.Repository) *Service {
	return &Service{
		roundRepo:      roundRepo,
		tournamentRepo: tournamentRepo,
	}
}

// CreateRound creates a new round for a tournament.
func (s *Service) CreateRound(ctx context.Context, req CreateRoundRequest) (*rounddomain.Round, error) {
	// Validate tournament exists
	_, err := s.tournamentRepo.GetByID(ctx, req.TournamentID)
	if err != nil {
		if errors.Is(err, tournamentdomain.ErrNotFound) {
			return nil, tournamentdomain.ErrNotFound
		}
		return nil, fmt.Errorf("failed to get tournament: %w", err)
	}

	// Validate round doesn't already exist
	existing, err := s.roundRepo.GetByTournamentAndNumber(ctx, req.TournamentID, req.Number)
	if err == nil && existing != nil {
		return nil, fmt.Errorf("round %d already exists for this tournament", req.Number)
	}
	if err != nil && !errors.Is(err, rounddomain.ErrNotFound) {
		return nil, fmt.Errorf("failed to check existing round: %w", err)
	}

	// Create round
	scheduledFor := time.Time{}
	if req.ScheduledFor != nil {
		scheduledFor = *req.ScheduledFor
	}

	round, err := rounddomain.NewRound(req.TournamentID, req.Number, scheduledFor)
	if err != nil {
		return nil, err
	}

	if req.MatchCount > 0 {
		round.SetMatchCount(req.MatchCount)
	}
	if req.Notes != "" {
		round.SetNotes(req.Notes)
	}

	// Persist round
	if err := s.roundRepo.Create(ctx, round); err != nil {
		return nil, fmt.Errorf("failed to create round: %w", err)
	}

	return round, nil
}

// GetRound retrieves a round by ID.
func (s *Service) GetRound(ctx context.Context, roundID uuid.UUID) (*rounddomain.Round, error) {
	round, err := s.roundRepo.GetByID(ctx, roundID)
	if err != nil {
		if errors.Is(err, rounddomain.ErrNotFound) {
			return nil, rounddomain.ErrNotFound
		}
		return nil, fmt.Errorf("failed to get round: %w", err)
	}
	return round, nil
}

// GetTournamentRounds retrieves all rounds for a tournament.
func (s *Service) GetTournamentRounds(ctx context.Context, tournamentID uuid.UUID, status *rounddomain.Status) ([]*rounddomain.Round, error) {
	rounds, err := s.roundRepo.GetByTournament(ctx, tournamentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get tournament rounds: %w", err)
	}

	// Filter by status if provided
	if status != nil {
		filtered := make([]*rounddomain.Round, 0)
		for _, r := range rounds {
			if r.Status == *status {
				filtered = append(filtered, r)
			}
		}
		return filtered, nil
	}

	return rounds, nil
}

// GetRoundByNumber retrieves a round by tournament and round number.
func (s *Service) GetRoundByNumber(ctx context.Context, tournamentID uuid.UUID, number int) (*rounddomain.Round, error) {
	round, err := s.roundRepo.GetByTournamentAndNumber(ctx, tournamentID, number)
	if err != nil {
		if errors.Is(err, rounddomain.ErrNotFound) {
			return nil, rounddomain.ErrNotFound
		}
		return nil, fmt.Errorf("failed to get round: %w", err)
	}
	return round, nil
}

// UpdateRound updates round details (schedule, match count, notes).
func (s *Service) UpdateRound(ctx context.Context, roundID uuid.UUID, req UpdateRoundRequest) (*rounddomain.Round, error) {
	round, err := s.roundRepo.GetByID(ctx, roundID)
	if err != nil {
		if errors.Is(err, rounddomain.ErrNotFound) {
			return nil, rounddomain.ErrNotFound
		}
		return nil, fmt.Errorf("failed to get round: %w", err)
	}

	// Update fields
	if req.ScheduledFor != nil {
		round.ScheduledFor = *req.ScheduledFor
	}
	if req.MatchCount != nil && *req.MatchCount > 0 {
		round.SetMatchCount(*req.MatchCount)
	}
	if req.Notes != nil {
		round.SetNotes(*req.Notes)
	}

	if err := s.roundRepo.Update(ctx, round); err != nil {
		return nil, fmt.Errorf("failed to update round: %w", err)
	}

	return round, nil
}

// UpdateRoundStatus updates round status (pending -> ongoing -> completed, or cancel from any state).
func (s *Service) UpdateRoundStatus(ctx context.Context, roundID uuid.UUID, req UpdateRoundStatusRequest) (*rounddomain.Round, error) {
	round, err := s.roundRepo.GetByID(ctx, roundID)
	if err != nil {
		if errors.Is(err, rounddomain.ErrNotFound) {
			return nil, rounddomain.ErrNotFound
		}
		return nil, fmt.Errorf("failed to get round: %w", err)
	}

	// Apply status transition
	switch req.Status {
	case rounddomain.StatusOngoing:
		if err := round.Start(); err != nil {
			return nil, fmt.Errorf("cannot start round: %w", err)
		}
	case rounddomain.StatusCompleted:
		if err := round.Complete(); err != nil {
			return nil, fmt.Errorf("cannot complete round: %w", err)
		}
	case rounddomain.StatusCanceled:
		if err := round.Cancel(); err != nil {
			return nil, fmt.Errorf("cannot cancel round: %w", err)
		}
	default:
		return nil, fmt.Errorf("invalid status transition: %s", req.Status)
	}

	if err := s.roundRepo.Update(ctx, round); err != nil {
		return nil, fmt.Errorf("failed to update round status: %w", err)
	}
	return round, nil
}

// DeleteRound deletes a round (only if pending and no matches yet).
func (s *Service) DeleteRound(ctx context.Context, roundID uuid.UUID) error {
	round, err := s.roundRepo.GetByID(ctx, roundID)
	if err != nil {
		if errors.Is(err, rounddomain.ErrNotFound) {
			return rounddomain.ErrNotFound
		}
		return fmt.Errorf("failed to get round: %w", err)
	}

	// Allow deletion only if pending
	if round.Status != rounddomain.StatusPending {
		return fmt.Errorf("can only delete pending rounds")
	}

	if err := s.roundRepo.Delete(ctx, roundID); err != nil {
		return fmt.Errorf("failed to delete round: %w", err)
	}

	return nil
}
