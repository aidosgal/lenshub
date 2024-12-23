package bot

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/aidosgal/lenshub/internal/model"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type UserService interface {
	CreateUser(model.User) error
    GetUserByChatID(chatID string) (*model.User, error)
    GetUsersBySpecialization(specialization string) ([]model.User, error)
}

type OrderService interface {
    CreateOrder(order model.Order) (model.Order, error)
    GetOrderByID(orderID string) (model.Order, error)
}

type OrderResponseService interface {
    CreateOrderResponse(orderID string, executorID int) error
}

type TgBot struct {
	bot        tgbotapi.BotAPI
	service    UserService
    orderService OrderService
    orderResponseService OrderResponseService
	userStates map[int64]string // Tracks the state of users
	userData   map[int64]*model.User
    orderData   map[int64]*model.Order
	stateMutex sync.Mutex
}

const (
	StateChoosingRole           = "choosing_role"
	StateEnteringPortfolio      = "entering_portfolio"
	StateChoosingSpecialization = "choosing_specialization"
	StateIdle                   = "idle"
    StateEnteringOrderTitle       = "entering_order_title"
    StateEnteringOrderDescription = "entering_order_description"
    StateEnteringOrderLocation    = "entering_order_location"
    StateChoosingOrderSpecialization = "choosing_order_specialization"
)

func NewTgBot(token string, service UserService, order OrderService, orderOrderResponseService OrderResponseService) *TgBot {
	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		panic(err)
	}
	return &TgBot{
		bot:        *bot,
		service:    service,
        orderService: order,
        orderResponseService: orderOrderResponseService,
		userStates: make(map[int64]string),
		userData:   make(map[int64]*model.User),
		orderData:   make(map[int64]*model.Order),
	}
}

func (tg *TgBot) Start() {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := tg.bot.GetUpdatesChan(u)

	for update := range updates {
        log.Printf("Received update: %+v\n", update)
		if update.Message != nil {
			tg.handleMessage(update.Message)
		}

		if update.CallbackQuery != nil {
			tg.handleCallbackQuery(update.CallbackQuery)
		}
	}
}

func (tg *TgBot) handleMessage(message *tgbotapi.Message) {
	chatID := message.Chat.ID
    chat_id := strconv.Itoa(int(chatID))

	tg.stateMutex.Lock()
	state := tg.userStates[chatID]
	tg.stateMutex.Unlock()

	if message.Text == "/start" {
        if user, err := tg.service.GetUserByChatID(chat_id); err == nil && user != nil {
            tg.showUserProfile(chatID, user)
            return
        }
		tg.stateMutex.Lock()
		tg.userStates[message.Chat.ID] = StateChoosingRole
		tg.stateMutex.Unlock()

        welcomeText := `👋 Добро пожаловать в LensHub!

Мы соединяем талантливых фотографов и видеографов с заказчиками.

Выберите вашу роль для начала работы:`

		buttons := [][]tgbotapi.InlineKeyboardButton{
            {
                tgbotapi.NewInlineKeyboardButtonData("🤝 Я заказчик", "role_customer"),
                tgbotapi.NewInlineKeyboardButtonData("📸 Я исполнитель", "role_executor"),
            },
        }
		keyboard := tgbotapi.NewInlineKeyboardMarkup(buttons...)

        response := tgbotapi.NewMessage(message.Chat.ID, welcomeText)
        response.ReplyMarkup = keyboard
        tg.bot.Send(response)
	}

	switch state {
	case StateEnteringPortfolio:
		tg.handlePortfolioInput(message)
    case StateEnteringOrderTitle:
        tg.handleOrderTitleInput(message)
    case StateEnteringOrderDescription:
        tg.handleOrderDescriptionInput(message)
    case StateEnteringOrderLocation:
        tg.handleOrderLocationInput(message)
	default:
		//response := tgbotapi.NewMessage(chatID, "Используйте /start для начала регистрации.")
		//tg.bot.Send(response)
	}
}

func (tg *TgBot) showUserProfile(chatID int64, user *model.User) {
    var profileText string
    var buttons [][]tgbotapi.InlineKeyboardButton

    if user.Role == "Заказчик" {
        profileText = fmt.Sprintf(`👤 *Ваш профиль*
        
📋 *Роль:* Заказчик
👤 *Имя:* %s
🔍 *Username:* @%s

Что бы вы хотели сделать?`, user.Name, user.UserName)

        buttons = [][]tgbotapi.InlineKeyboardButton{
            {
                tgbotapi.NewInlineKeyboardButtonData("📝 Создать заказ", "create_order"),
                tgbotapi.NewInlineKeyboardButtonData("📋 Мои заказы", "my_orders"),
            },
        }
    } else {
        profileText = fmt.Sprintf(`👤 *Ваш профиль*
        
📸 *Роль:* Исполнитель
👤 *Имя:* %s
🔍 *Username:* @%s
🎯 *Специализация:* %s

Что бы вы хотели сделать?`, user.Name, user.UserName, user.Specialization)

        buttons = [][]tgbotapi.InlineKeyboardButton{
            {
                tgbotapi.NewInlineKeyboardButtonURL("🎨 Моё портфолио", user.Portfolio),
            },
        }
    }

    keyboard := tgbotapi.NewInlineKeyboardMarkup(buttons...)
    msg := tgbotapi.NewMessage(chatID, profileText)
    msg.ParseMode = "Markdown"
    msg.ReplyMarkup = keyboard
    tg.bot.Send(msg)
}

func (tg *TgBot) handleCallbackQuery(callbackQuery *tgbotapi.CallbackQuery) {
    chatID := callbackQuery.Message.Chat.ID
    data := callbackQuery.Data
    log.Printf("Handling callback query with data: %s for chatID: %d", data, chatID)

    // Create a copy of userData before locking
    var userData *model.User
    chat_id := strconv.Itoa(int(chatID))

    tg.stateMutex.Lock()
    if existing, ok := tg.userData[chatID]; ok {
        userData = &model.User{
            Name:           callbackQuery.From.FirstName,
            UserName:           callbackQuery.From.UserName,
            Role:           existing.Role,
            ChatId:        chat_id,
            Portfolio:      existing.Portfolio,
            Specialization: existing.Specialization,
        }
    } else {
        userData = &model.User{ChatId: chat_id}
    }
    tg.stateMutex.Unlock()

    switch { 
    case data == "role_customer":
        userData.Role = "Заказчик"
        userData.Name = callbackQuery.From.FirstName
        userData.UserName = callbackQuery.From.UserName
        if err := tg.service.CreateUser(*userData); err != nil {
            log.Printf("Error creating customer: %v", err)
            response := tgbotapi.NewMessage(chatID, "❌ Произошла ошибка при регистрации. Пожалуйста, попробуйте еще раз.")
            tg.bot.Send(response)
            return
        }
        
        tg.stateMutex.Lock()
        tg.userStates[chatID] = StateIdle
        tg.userData[chatID] = userData
        tg.stateMutex.Unlock()

        successMsg := fmt.Sprintf(`✅ Регистрация успешно завершена!

🤝 Добро пожаловать в команду заказчиков, %s!

Теперь вы можете:
- Создавать новые заказы
- Просматривать отклики исполнителей
- Управлять своими заказами`, userData.Name)

        response := tgbotapi.NewMessage(chatID, successMsg)
        tg.bot.Send(response)
        tg.showUserProfile(chatID, userData)

    case data == "role_executor":
        userData.Role = "Исполнитель"
        
        portfolioMsg := `📸 Для завершения регистрации, пожалуйста, отправьте ссылку на ваше портфолио.

Это может быть:
- Ссылка на Instagram
- Ссылка на личный сайт
- Ссылка на облачное хранилище с работами
- Любой другой ресурс с вашими работами`

        tg.stateMutex.Lock()
        tg.userStates[chatID] = StateEnteringPortfolio
        tg.userData[chatID] = userData
        tg.stateMutex.Unlock()

        response := tgbotapi.NewMessage(chatID, portfolioMsg)
        tg.bot.Send(response)

    case data == "specialization_videographer" || data == "specialization_photographer":
        userData.Specialization = map[string]string{
            "specialization_videographer": "Видеооператор",
            "specialization_photographer": "Фотограф",
        }[data]
        
        go tg.completeExecutorRegistration(chatID, userData)    
    case data == "create_order":
        tg.startOrderCreation(chatID)
    case strings.HasPrefix(data, "respond_to_order:"):
        orderID := strings.Split(data, ":")[1]
        tg.handleOrderResponse(chatID, orderID)
    case data == "order_spec_videographer" || data == "order_spec_photographer":
        tg.handleOrderSpecializationSelection(chatID, data)
    }
}

func (tg *TgBot) startOrderCreation(chatID int64) {
    tg.stateMutex.Lock()
    tg.orderData[chatID] = &model.Order{}
    tg.userStates[chatID] = StateChoosingOrderSpecialization
    tg.stateMutex.Unlock()

    specMsg := `🎯 Выберите тип специалиста для вашего заказа:

- Видеооператор - съемка и монтаж видео
- Фотограф - фотосъемка и обработка фото`

    buttons := [][]tgbotapi.InlineKeyboardButton{
        {
            tgbotapi.NewInlineKeyboardButtonData("🎥 Видеооператор", "order_spec_videographer"),
            tgbotapi.NewInlineKeyboardButtonData("📸 Фотограф", "order_spec_photographer"),
        },
    }
    keyboard := tgbotapi.NewInlineKeyboardMarkup(buttons...)

    response := tgbotapi.NewMessage(chatID, specMsg)
    response.ReplyMarkup = keyboard
    tg.bot.Send(response)
}


func (tg *TgBot) completeExecutorRegistration(chatID int64, userData *model.User) {
    log.Printf("Starting executor registration for chatID: %d with data: %+v", chatID, userData)
    
    err := tg.service.CreateUser(*userData)
    if err != nil {
        log.Printf("Error creating executor: %v", err)
        errorMsg := `❌ Произошла ошибка при регистрации. 

Пожалуйста, попробуйте еще раз или свяжитесь с поддержкой.`
        response := tgbotapi.NewMessage(chatID, errorMsg)
        if _, err := tg.bot.Send(response); err != nil {
            log.Printf("Error sending error message: %v", err)
        }
        return
    }
    log.Println("User created successfully in database")

    tg.stateMutex.Lock()
    tg.userStates[chatID] = StateIdle
    tg.userData[chatID] = userData
    tg.stateMutex.Unlock()
    log.Println("User state updated to idle")

    successMsg := fmt.Sprintf(`✅ Регистрация успешно завершена!

🎨 Добро пожаловать в команду исполнителей, %s!

Ваша специализация: %s

Теперь вы можете:
- Просматривать доступные заказы
- Откликаться на интересные проекты
- Общаться с заказчиками`, userData.Name, userData.Specialization)

    response := tgbotapi.NewMessage(chatID, successMsg)
    if _, err := tg.bot.Send(response); err != nil {
        log.Printf("Error sending confirmation: %v", err)
        return
    }
    
    tg.showUserProfile(chatID, userData)
    
    log.Printf("Registration completed successfully for chatID: %d", chatID)
}

func (tg *TgBot) handlePortfolioInput(message *tgbotapi.Message) {
    chatID := message.Chat.ID
    tg.userData[chatID].Portfolio = message.Text

    tg.stateMutex.Lock()
    tg.userStates[chatID] = StateChoosingSpecialization
    tg.stateMutex.Unlock()

    specMsg := `🎯 Выберите вашу специализацию:

- Видеооператор - съемка и монтаж видео
- Фотограф - фотосъемка и обработка фото`

    buttons := [][]tgbotapi.InlineKeyboardButton{
        {
            tgbotapi.NewInlineKeyboardButtonData("🎥 Видеооператор", "specialization_videographer"),
            tgbotapi.NewInlineKeyboardButtonData("📸 Фотограф", "specialization_photographer"),
        },
    }
    keyboard := tgbotapi.NewInlineKeyboardMarkup(buttons...)

    response := tgbotapi.NewMessage(chatID, specMsg)
    response.ReplyMarkup = keyboard
    tg.bot.Send(response)
}

func (tg *TgBot) handleOrderTitleInput(message *tgbotapi.Message) {
    chatID := message.Chat.ID
    
    tg.stateMutex.Lock()
    tg.orderData[chatID].Title = message.Text
    tg.userStates[chatID] = StateEnteringOrderDescription
    tg.stateMutex.Unlock()

    msg := `📝 Отлично! Теперь опишите подробности заказа:

Например:
- Что конкретно нужно сделать
- Когда планируется съемка
- Особые пожелания или требования`

    response := tgbotapi.NewMessage(chatID, msg)
    tg.bot.Send(response)
}

func (tg *TgBot) handleOrderDescriptionInput(message *tgbotapi.Message) {
    chatID := message.Chat.ID
    
    tg.stateMutex.Lock()
    tg.orderData[chatID].Description = message.Text
    tg.userStates[chatID] = StateEnteringOrderLocation
    tg.stateMutex.Unlock()

    msg := `📍 Укажите место проведения съемки:

Например: "Алматы, парк Горького" или "Студия на Абая 150"`

    response := tgbotapi.NewMessage(chatID, msg)
    tg.bot.Send(response)
}

func (tg *TgBot) handleOrderLocationInput(message *tgbotapi.Message) {
    chatID := message.Chat.ID
    
    tg.stateMutex.Lock()
    order := tg.orderData[chatID]
    order.Location = message.Text
    order.CreatedAt = time.Now()
    user, err := tg.service.GetUserByChatID(strconv.FormatInt(chatID, 10))
    if err != nil {
        log.Printf("Error getting user: %v", err)
        response := tgbotapi.NewMessage(chatID, "❌ Произошла ошибка при создании заказа. Попробуйте еще раз.")
        tg.bot.Send(response)

        return
    }
    order.User = *user
    tg.userStates[chatID] = StateIdle
    defer tg.stateMutex.Unlock()

    // Create order in database
    createdOrder, err := tg.orderService.CreateOrder(*order)
    if err != nil {
        log.Printf("Error creating order: %v", err)
        response := tgbotapi.NewMessage(chatID, "❌ Произошла ошибка при создании заказа. Попробуйте еще раз.")
        tg.bot.Send(response)
        return
    }

    // Notify user about successful creation
    successMsg := fmt.Sprintf(`✅ Заказ успешно создан!

📋 *%s*
📝 %s
📍 %s

Мы уведомим исполнителей о вашем заказе.`, order.Title, order.Description, order.Location)

    response := tgbotapi.NewMessage(chatID, successMsg)
    response.ParseMode = "Markdown"
    tg.bot.Send(response)

    // Notify potential executors
    tg.notifyExecutors(&createdOrder)
}

func (tg *TgBot) notifyExecutors(order *model.Order) {
    executors, err := tg.service.GetUsersBySpecialization(order.Specialization)
    if err != nil {
        log.Printf("Error getting executors: %v", err)
        return
    }

    log.Println(executors)
    for _, executor := range executors {
        chatID, _ := strconv.ParseInt(executor.ChatId, 10, 64)
        
        notificationMsg := fmt.Sprintf(`🆕 Новый заказ!

📋 *%s*
🎯 Специализация: *%s*
📝 %s
📍 %s
🕒 %s

Заинтересованы в этом заказе?`, 
            order.Title,
            order.Specialization,
            order.Description,
            order.Location,
            order.CreatedAt.Format("02.01.2006 15:04"))

        buttons := [][]tgbotapi.InlineKeyboardButton{
            {
                tgbotapi.NewInlineKeyboardButtonData("✅ Откликнуться", fmt.Sprintf("respond_to_order:%d", order.ID)),
            },
        }
        keyboard := tgbotapi.NewInlineKeyboardMarkup(buttons...)

        msg := tgbotapi.NewMessage(chatID, notificationMsg)
        msg.ParseMode = "Markdown"
        msg.ReplyMarkup = keyboard
        
        if _, err := tg.bot.Send(msg); err != nil {
            log.Printf("Error notifying executor %d: %v", chatID, err)
        }
    }
}

func (tg *TgBot) handleOrderResponse(chatID int64, orderID string) {
    log.Printf("Starting to handle order response for chatID: %d, orderID: %s", chatID, orderID)

    // Get executor info
    log.Printf("Fetching executor info for chatID: %d", chatID)
    executor, err := tg.service.GetUserByChatID(strconv.FormatInt(chatID, 10))
    if err != nil {
        log.Printf("Error getting executor with chatID %d: %v", chatID, err)
        response := tgbotapi.NewMessage(chatID, "❌ Произошла ошибка. Пожалуйста, попробуйте позже.")
        tg.bot.Send(response)
        return
    }
    log.Printf("Successfully retrieved executor: %+v", executor)

    // Store response in database
    log.Printf("Creating order response in database. OrderID: %s, ExecutorID: %s", orderID, executor.ChatId)
    if err = tg.orderResponseService.CreateOrderResponse(orderID, executor.Id); err != nil {
        log.Printf("Error creating order response in database: %v", err)
        response := tgbotapi.NewMessage(chatID, "❌ Произошла ошибка при сохранении отклика. Пожалуйста, попробуйте позже.")
        tg.bot.Send(response)
        return
    }
    log.Printf("Successfully created order response in database")

    // Send confirmation to executor
    log.Printf("Sending confirmation message to executor with chatID: %d", chatID)
    msg := `✅ Вы успешно откликнулись на заказ!

Заказчик получит уведомление с вашим профилем и свяжется с вами через Telegram.`
    response := tgbotapi.NewMessage(chatID, msg)
    if _, err := tg.bot.Send(response); err != nil {
        log.Printf("Error sending confirmation to executor: %v", err)
        return
    }
    log.Printf("Successfully sent confirmation to executor")

    // Get order details to find customer
    log.Printf("Fetching order details for orderID: %s", orderID)
    order, err := tg.orderService.GetOrderByID(orderID)
    if err != nil {
        log.Printf("Error getting order details: %v", err)
        return
    }
    log.Printf("Successfully retrieved order: %+v", order)

    // Convert customer's chat ID from string to int64
    log.Printf("Converting customer chatID from string to int64: %s", order.User.ChatId)
    customerChatID, err := strconv.ParseInt(order.User.ChatId, 10, 64)
    if err != nil {
        log.Printf("Error parsing customer chat ID '%s': %v", order.User.ChatId, err)
        return
    }
    log.Printf("Successfully converted customer chatID to: %d", customerChatID)

    // Escape the username to handle Telegram Markdown formatting
    escapedUserName := escapeMarkdown(executor.UserName)

    // Prepare and send executor's profile to customer
    log.Printf("Preparing notification message for customer with chatID: %d", customerChatID)
    profileText := fmt.Sprintf(`🔔 Новый отклик на ваш заказ *"%s"*, теперь вы можете связаться с исполнителем напрямую!

👤 *Профиль исполнителя:*
📸 *Роль:* %s
👤 *Имя:* %s
🔍 *Username:* @%s
🎯 *Специализация:* %s`, 
        escapeMarkdown(order.Title),
        escapeMarkdown(executor.Role),
        escapeMarkdown(executor.Name),
        escapedUserName,
        escapeMarkdown(executor.Specialization),
    )

    buttons := [][]tgbotapi.InlineKeyboardButton{
        {
            tgbotapi.NewInlineKeyboardButtonURL("🎨 Портфолио исполнителя", executor.Portfolio),
        },
    }
    keyboard := tgbotapi.NewInlineKeyboardMarkup(buttons...)

    customerMsg := tgbotapi.NewMessage(customerChatID, profileText)
    customerMsg.ParseMode = "Markdown"
    customerMsg.ReplyMarkup = keyboard

    log.Printf("Sending notification to customer with chatID: %d", customerChatID)
    if _, err := tg.bot.Send(customerMsg); err != nil {
        log.Printf("Error sending notification to customer: %v", err)
        return
    }
    log.Printf("Successfully sent notification to customer about new response")
}


func (tg *TgBot) handleOrderSpecializationSelection(chatID int64, data string) {
    tg.stateMutex.Lock()
    order, exists := tg.orderData[chatID]
    if !exists {
        tg.stateMutex.Unlock()
        response := tgbotapi.NewMessage(chatID, "❌ Произошла ошибка. Пожалуйста, начните создание заказа заново.")
        tg.bot.Send(response)
        return
    }

    // Set specialization based on selection
    order.Specialization = map[string]string{
        "order_spec_videographer": "Видеооператор",
        "order_spec_photographer": "Фотограф",
    }[data]

    // Move to next state
    tg.userStates[chatID] = StateEnteringOrderTitle
    tg.stateMutex.Unlock()

    // Prompt for order title
    msg := `📝 Отлично! Теперь введите название заказа:
Например: "Свадебная фотосессия" или "Видеосъёмка дня рождения"`

    response := tgbotapi.NewMessage(chatID, msg)
    tg.bot.Send(response)
}

func escapeMarkdown(text string) string {
    // Escape Telegram Markdown special characters
    replacer := strings.NewReplacer(
        "_", "\\_",
        "*", "\\*",
        "[", "\\[",
        "]", "\\]",
        "(", "\\(",
        ")", "\\)",
        "`", "\\`",
    )
    return replacer.Replace(text)
}
