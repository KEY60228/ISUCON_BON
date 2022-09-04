package main

import (
	"context"
	"math/rand"
	"sync"

	"github.com/isucon/isucandar"
	"github.com/isucon/isucandar/failure"
	"github.com/isucon/isucandar/score"
	"github.com/isucon/isucandar/worker"
)

const (
	ErrFailedLoadJSON  failure.StringCode = "load-json"
	ErrCannotNewAgent  failure.StringCode = "agent"
	ErrInvalidRequest  failure.StringCode = "request"
	ErrInvalidResposne failure.StringCode = "response"
)

const (
	ScoreGETLogin  score.ScoreTag = "GET /login"
	ScorePOSTLogin score.ScoreTag = "POST /login"
	ScoreGETRoot   score.ScoreTag = "GET /"
	ScorePOSTRoot  score.ScoreTag = "POST /"
)

type Scenario struct {
	Option   Option
	Users    UserSet
	Posts    PostSet
	Comments CommentSet
}

func (s *Scenario) Prepare(ctx context.Context, step *isucandar.BenchmarkStep) error {
	if err := s.Users.LoadJSON("./dump/users.json"); err != nil {
		return failure.NewError(ErrFailedLoadJSON, err)
	}

	if err := s.Posts.LoadJSON("./dump/posts.json"); err != nil {
		return failure.NewError(ErrFailedLoadJSON, err)
	}

	if err := s.Comments.LoadJSON("./dump/commnets.json"); err != nil {
		return failure.NewError(ErrFailedLoadJSON, err)
	}

	ag, err := s.Option.NewAgent(true)
	if err != nil {
		return failure.NewError(ErrCannotNewAgent, err)
	}

	res, err := GetInitializeAction(ctx, ag)
	if err != nil {
		return failure.NewError(ErrInvalidRequest, err)
	}
	defer res.Body.Close()

	ValidateResponse(res, WithStatusCode(200)).Add(step)

	return nil
}

func (s *Scenario) Load(ctx context.Context, step *isucandar.BenchmarkStep) error {
	wg := &sync.WaitGroup{}

	successCase, err := worker.NewWorker(func(ctx context.Context, _ int) {
		if user, ok := s.Users.Get(rand.Intn(s.Users.Len())); ok {
			if user.DeleteFlag != 0 {
				return
			}

			if s.LoginSuccess(ctx, step, user) {
				s.PostImage(ctx, step, user)
			}
			user.ClearAgent()
		}
	}, worker.WithInfinityLoop(), worker.WithMaxParallelism(4))
	if err != nil {
		return err
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
		successCase.Process(ctx)
	}()

	failureCase, err := worker.NewWorker(func(ctx context.Context, _ int) {
		if user, ok := s.Users.Get(rand.Intn(s.Users.Len())); ok {
			if user.DeleteFlag != 0 {
				return
			}
			s.LoginFailure(ctx, step, user)
		}
	}, worker.WithLoopCount(20), worker.WithMaxParallelism(2))
	if err != nil {
		return err
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
		failureCase.Process(ctx)
	}()

	orderedCase, err := worker.NewWorker(func(ctx context.Context, _ int) {
		if user, ok := s.Users.Get(rand.Intn(s.Users.Len())); ok {
			s.OrderedIndex(ctx, step, user)
		}
	}, worker.WithInfinityLoop(), worker.WithMaxParallelism(2))
	if err != nil {
		return err
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
		orderedCase.Process(ctx)
	}()

	wg.Wait()
	return nil
}

func (s *Scenario) LoginSuccess(ctx context.Context, step *isucandar.BenchmarkStep, user *User) bool {
	ag, err := user.GetAgent(s.Option)
	if err != nil {
		step.AddError(failure.NewError(ErrCannotNewAgent, err))
		return false
	}

	getRes, err := GetLoginAction(ctx, ag)
	if err != nil {
		step.AddError(failure.NewError(ErrInvalidRequest, err))
		return false
	}
	defer getRes.Body.Close()

	getValidation := ValidateResponse(getRes, WithStatusCode(200), WithAssets(ctx, ag))
	getValidation.Add(step)

	if getValidation.IsEmpty() {
		step.AddScore(ScoreGETLogin)
	} else {
		return false
	}

	select {
	case <-ctx.Done():
		return false
	default:
	}

	postRes, err := PostLoginAction(ctx, ag, user.AccountName, user.Password)
	if err != nil {
		step.AddError(failure.NewError(ErrInvalidRequest, err))
		return false
	}
	defer postRes.Body.Close()

	postValidation := ValidateResponse(postRes, WithStatusCode(302), WithLocation("/"))

	if postValidation.IsEmpty() {
		step.AddScore(ScorePOSTLogin)
	} else {
		return false
	}

	return true
}

func (s *Scenario) LoginFailure(ctx context.Context, step *isucandar.BenchmarkStep, user *User) bool {
	ag, err := user.GetAgent(s.Option)
	if err != nil {
		step.AddError(failure.NewError(ErrCannotNewAgent, err))
		return false
	}

	getRes, err := GetLoginAction(ctx, ag)
	if err != nil {
		step.AddError(failure.NewError(ErrInvalidRequest, err))
		return false
	}
	defer getRes.Body.Close()

	getValidation := ValidateResponse(getRes, WithStatusCode(200), WithAssets(ctx, ag))
	getValidation.Add(step)

	if getValidation.IsEmpty() {
		step.AddScore(ScoreGETLogin)
	} else {
		return false
	}

	select {
	case <-ctx.Done():
		return false
	default:
	}

	postRes, err := PostLoginAction(ctx, ag, user.AccountName, user.Password+".invalid")
	if err != nil {
		step.AddError(failure.NewError(ErrInvalidRequest, err))
		return false
	}
	defer postRes.Body.Close()

	postValidation := ValidateResponse(postRes, WithStatusCode(302), WithLocation("/login"))
	postValidation.Add(step)

	if postValidation.IsEmpty() {
		step.AddScore(ScorePOSTLogin)
	} else {
		return false
	}

	select {
	case <-ctx.Done():
		return false
	default:
	}

	redirectRes, err := GetLoginAction(ctx, ag)
	if err != nil {
		step.AddError(failure.NewError(ErrInvalidRequest, err))
		return false
	}
	defer getRes.Body.Close()

	redirectValidation := ValidateResponse(redirectRes, WithStatusCode(200), WithIncludeBody("アカウント名かパスワードが間違っています"))
	redirectValidation.Add(step)

	if redirectValidation.IsEmpty() {
		step.AddScore(ScoreGETLogin)
	} else {
		return false
	}

	return true
}

func (s *Scenario) PostImage(ctx context.Context, step *isucandar.BenchmarkStep, user *User) bool {
	ag, err := user.GetAgent(s.Option)
	if err != nil {
		step.AddError(failure.NewError(ErrCannotNewAgent, err))
		return false
	}

	getRes, err := GetRootAction(ctx, ag)
	if err != nil {
		step.AddError(failure.NewError(ErrInvalidRequest, err))
		return false
	}
	defer getRes.Body.Close()

	getValidation := ValidateResponse(getRes, WithStatusCode(200), WithCSRFToken(user))
	getValidation.Add(step)

	if getValidation.IsEmpty() {
		step.AddScore(ScoreGETRoot)
	} else {
		return false
	}

	select {
	case <-ctx.Done():
		return false
	default:
	}

	post := &Post{
		Mime:   "image/png",
		Body:   randomText(),
		UserID: user.ID,
	}
	postRes, err := PostRootAction(ctx, ag, post, user.GetCSRFToken())
	if err != nil {
		step.AddError(failure.NewError(ErrInvalidRequest, err))
		return false
	}
	defer postRes.Body.Close()

	postValidation := ValidateResponse(postRes, WithStatusCode(302))
	postValidation.Add(step)

	if postValidation.IsEmpty() {
		step.AddScore(ScorePOSTRoot)
	} else {
		return false
	}

	select {
	case <-ctx.Done():
		return false
	default:
	}

	redirectRes, err := GetRootAction(ctx, ag)
	if err != nil {
		step.AddError(failure.NewError(ErrInvalidRequest, err))
		return false
	}
	defer getRes.Body.Close()

	redirectValidation := ValidateResponse(redirectRes, WithStatusCode(200), WithAssets(ctx, ag))
	redirectValidation.Add(step)

	if redirectValidation.IsEmpty() {
		step.AddScore(ScoreGETRoot)
	} else {
		return false
	}

	return true
}
