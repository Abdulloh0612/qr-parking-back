package bot

import (
	"context"
	"fmt"
	"time"

	"qr-parking/db/repositories"
	"qr-parking/types"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	tele "gopkg.in/telebot.v3"
)

type Bot struct {
	bot      *tele.Bot
	tgRepo   repositories.TelegramRepository
	userRepo repositories.UserRepository
	redis    *redis.Client
	logger   *zap.Logger
}

func New(token string, tgRepo repositories.TelegramRepository, userRepo repositories.UserRepository, redisClient *redis.Client, logger *zap.Logger) (*Bot, error) {
	pref := tele.Settings{
		Token:  token,
		Poller: &tele.LongPoller{Timeout: 10 * time.Second},
	}

	b, err := tele.NewBot(pref)
	if err != nil {
		return nil, fmt.Errorf("create bot: %w", err)
	}

	return &Bot{
		bot:      b,
		tgRepo:   tgRepo,
		userRepo: userRepo,
		redis:    redisClient,
		logger:   logger,
	}, nil
}

func (b *Bot) Start() {
	b.bot.Handle("/start", b.handleStart)
	b.bot.Handle("/status", b.handleStatus)
	b.bot.Handle("/stop", b.handleStop)
	b.bot.Handle("/help", b.handleHelp)

	b.logger.Info("Telegram bot started")
	b.bot.Start()
}

func (b *Bot) Stop() {
	b.bot.Stop()
}

func (b *Bot) handleStart(c tele.Context) error {
	payload := c.Message().Payload
	if payload == "" {
		return c.Send("Добро пожаловать в QR-Parking Bot!\n\nДля привязки аккаунта перейдите в настройки профиля на сайте и нажмите 'Привязать Telegram'.\n\n/help — справка")
	}

	ctx := context.Background()
	key := fmt.Sprintf("tg:link:%s", payload)
	userIDStr, err := b.redis.Get(ctx, key).Result()
	if err != nil {
		return c.Send("Ссылка для привязки недействительна или истекла. Попробуйте снова через настройки профиля.")
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return c.Send("Ошибка привязки. Попробуйте снова.")
	}

	user, err := b.userRepo.GetByID(ctx, userID)
	if err != nil || user == nil {
		return c.Send("Пользователь не найден.")
	}

	tgUsername := c.Sender().Username
	acc := &types.TelegramAccount{
		ID:         uuid.New(),
		UserID:     userID,
		TgUserID:   c.Sender().ID,
		TgUsername: &tgUsername,
	}

	if err := b.tgRepo.Create(ctx, acc); err != nil {
		b.logger.Error("failed to link telegram", zap.Error(err))
		return c.Send("Ошибка при привязке аккаунта. Попробуйте позже.")
	}

	b.redis.Del(ctx, key)

	return c.Send(fmt.Sprintf("Аккаунт успешно привязан!\n\nПривет, %s! Теперь вы будете получать уведомления о сканированиях вашего QR-кода.", user.FirstName))
}

func (b *Bot) handleStatus(c tele.Context) error {
	ctx := context.Background()
	acc, err := b.tgRepo.GetByTgUserID(ctx, c.Sender().ID)
	if err != nil || acc == nil {
		return c.Send("Ваш Telegram не привязан к аккаунту QR-Parking.\n\nПерейдите в настройки профиля на сайте для привязки.")
	}

	user, _ := b.userRepo.GetByID(ctx, acc.UserID)
	if user != nil {
		return c.Send(fmt.Sprintf("Аккаунт привязан.\nИмя: %s %s\nПривязан: %s", user.FirstName, user.LastName, acc.LinkedAt.Format("02.01.2006 15:04")))
	}

	return c.Send("Аккаунт привязан.")
}

func (b *Bot) handleStop(c tele.Context) error {
	ctx := context.Background()
	acc, err := b.tgRepo.GetByTgUserID(ctx, c.Sender().ID)
	if err != nil || acc == nil {
		return c.Send("Ваш аккаунт не привязан.")
	}

	if err := b.tgRepo.Delete(ctx, acc.UserID); err != nil {
		return c.Send("Ошибка при отвязке. Попробуйте позже.")
	}

	return c.Send("Аккаунт отвязан от уведомлений. Вы больше не будете получать сообщения о сканированиях.")
}

func (b *Bot) handleHelp(c tele.Context) error {
	return c.Send("QR-Parking Bot — уведомления о сканированиях вашего QR-кода.\n\nКоманды:\n/start TOKEN — привязать аккаунт\n/status — статус привязки\n/stop — отвязать аккаунт\n/help — справка")
}

func (b *Bot) SendNotification(tgUserID int64, text string) error {
	_, err := b.bot.Send(&tele.User{ID: tgUserID}, text)
	return err
}
