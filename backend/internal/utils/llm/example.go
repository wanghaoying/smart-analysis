package llm

import (
	"context"
	"fmt"
	"log"
	"smart-analysis/internal/config"
)

// exampleUsage 展示如何使用LLM管理器进行聊天
func exampleUsage() {
	// 1. 加载应用配置
	appConfig := config.Load()

	// 2. 初始化全局管理器
	if err := InitializeGlobalManager(appConfig); err != nil {
		log.Fatalf("Failed to initialize LLM manager: %v", err)
	}

	// 3. 获取管理器实例
	manager := GetGlobalManager()

	// 4. 检查可用的提供商
	providers := manager.ListProviders()
	fmt.Printf("Available providers: %v\n", providers)

	// 5. 创建聊天请求
	req := &ChatRequest{
		Messages: []Message{
			{Role: "system", Content: "You are a helpful assistant."},
			{Role: "user", Content: "Hello, how are you?"},
		},
		Model:       "", // 使用默认模型
		MaxTokens:   100,
		Temperature: 0.7,
	}

	ctx := context.Background()

	// 6. 使用默认客户端进行阻塞式聊天
	fmt.Println("\n=== Blocking Chat Example ===")
	resp, err := manager.ChatWithDefault(ctx, req)
	if err != nil {
		log.Printf("Chat error: %v", err)
	} else {
		if resp.Error != nil {
			log.Printf("API error: %s", resp.Error.Message)
		} else if len(resp.Choices) > 0 {
			fmt.Printf("Response: %s\n", resp.Choices[0].Message.Content)
		}
	}

	// 7. 使用默认客户端进行流式聊天
	fmt.Println("\n=== Streaming Chat Example ===")
	eventChan, err := manager.StreamChatWithDefault(ctx, req)
	if err != nil {
		log.Printf("Stream chat error: %v", err)
		return
	}

	fmt.Print("Streaming response: ")
	for event := range eventChan {
		if event.Error != nil {
			log.Printf("Stream error: %v", event.Error)
			break
		}

		if event.Data != nil {
			if event.Data.Error != nil {
				log.Printf("API error: %s", event.Data.Error.Message)
				break
			}

			if len(event.Data.Choices) > 0 && event.Data.Choices[0].Delta != nil {
				fmt.Print(event.Data.Choices[0].Delta.Content)
			}
		}

		if event.Done {
			fmt.Println("\n[Stream completed]")
			break
		}
	}

	// 8. 使用特定提供商
	if len(providers) > 0 {
		fmt.Printf("\n=== Using specific provider: %s ===\n", providers[0])
		resp, err := manager.Chat(ctx, providers[0], req)
		if err != nil {
			log.Printf("Provider chat error: %v", err)
		} else {
			if resp.Error != nil {
				log.Printf("API error: %s", resp.Error.Message)
			} else if len(resp.Choices) > 0 {
				fmt.Printf("Response: %s\n", resp.Choices[0].Message.Content)
			}
		}
	}
}

// exampleManualSetup 展示手动设置客户端的方法
func exampleManualSetup() {
	// 创建管理器
	manager := NewClientManager()

	// 手动配置OpenAI
	openaiConfig := &Config{
		Provider:    ProviderOpenAI,
		APIKey:      "your-openai-api-key",
		BaseURL:     "https://api.openai.com/v1",
		Model:       "gpt-3.5-turbo",
		MaxTokens:   1000,
		Temperature: 0.7,
		Timeout:     60,
	}

	if err := manager.RegisterProvider(ProviderOpenAI, openaiConfig); err != nil {
		log.Printf("Failed to register OpenAI: %v", err)
	}

	// 手动配置混元
	hunyuanConfig := &Config{
		Provider:    ProviderHunyuan,
		APIKey:      "secretId:secretKey", // 替换为实际的密钥
		BaseURL:     "https://hunyuan.tencentcloudapi.com",
		Model:       "hunyuan-lite",
		MaxTokens:   1000,
		Temperature: 0.7,
		Timeout:     60,
	}

	if err := manager.RegisterProvider(ProviderHunyuan, hunyuanConfig); err != nil {
		log.Printf("Failed to register Hunyuan: %v", err)
	}

	// 使用客户端
	ctx := context.Background()
	req := &ChatRequest{
		Messages: []Message{
			{Role: "user", Content: "Hello!"},
		},
	}

	// 测试OpenAI
	if resp, err := manager.Chat(ctx, ProviderOpenAI, req); err != nil {
		log.Printf("OpenAI error: %v", err)
	} else {
		fmt.Printf("OpenAI response: %s\n", resp.Choices[0].Message.Content)
	}

	// 测试混元
	if resp, err := manager.Chat(ctx, ProviderHunyuan, req); err != nil {
		log.Printf("Hunyuan error: %v", err)
	} else {
		fmt.Printf("Hunyuan response: %s\n", resp.Choices[0].Message.Content)
	}

	// 清理资源
	if err := manager.Close(); err != nil {
		log.Printf("Failed to close manager: %v", err)
	}
}

// exampleWithDifferentConfigs 展示不同配置的使用
func exampleWithDifferentConfigs() {
	manager := NewClientManager()

	// 高创造性配置
	creativeConfig := &Config{
		Provider:    ProviderOpenAI,
		APIKey:      "your-api-key",
		Model:       "gpt-4",
		MaxTokens:   2000,
		Temperature: 1.2, // 高创造性
		Timeout:     120,
	}

	// 保守配置
	conservativeConfig := &Config{
		Provider:    ProviderOpenAI,
		APIKey:      "your-api-key",
		Model:       "gpt-3.5-turbo",
		MaxTokens:   500,
		Temperature: 0.1, // 低创造性
		Timeout:     30,
	}

	// 注册不同配置（注意：实际使用中同一provider只能注册一次）
	_ = creativeConfig
	_ = conservativeConfig
	_ = manager

	fmt.Println("Different configurations can be used for different scenarios")
}

func example() {
	// 创建客户端管理器
	manager := NewClientManager()

	// 配置OpenAI（如果有API Key的话）
	if apiKey := "your-openai-api-key"; apiKey != "" && apiKey != "your-openai-api-key" {
		openaiConfig := &Config{
			Provider:    ProviderOpenAI,
			APIKey:      apiKey,
			Model:       "gpt-3.5-turbo",
			MaxTokens:   1000,
			Temperature: 0.7,
			Timeout:     60,
		}

		if err := manager.RegisterProvider(ProviderOpenAI, openaiConfig); err != nil {
			log.Printf("Failed to register OpenAI: %v", err)
		} else {
			fmt.Println("✅ OpenAI client registered successfully")
		}
	}

	// 配置混元（如果有API Key的话）
	if apiKey := "your-secret-id:your-secret-key"; apiKey != "" && apiKey != "your-secret-id:your-secret-key" {
		hunyuanConfig := &Config{
			Provider:    ProviderHunyuan,
			APIKey:      apiKey,
			Model:       "hunyuan-lite",
			MaxTokens:   1000,
			Temperature: 0.7,
			Timeout:     60,
		}

		if err := manager.RegisterProvider(ProviderHunyuan, hunyuanConfig); err != nil {
			log.Printf("Failed to register Hunyuan: %v", err)
		} else {
			fmt.Println("✅ Hunyuan client registered successfully")
		}
	}

	// 检查注册的提供商
	providers := manager.ListProviders()
	if len(providers) == 0 {
		fmt.Println("⚠️ No LLM providers configured. Please set up API keys.")
		fmt.Println("   OpenAI: Set OPENAI_API_KEY environment variable")
		fmt.Println("   Hunyuan: Set HUNYUAN_API_KEY environment variable (format: secretId:secretKey)")
		return
	}

	fmt.Printf("📋 Available providers: %v\n", providers)

	// 创建测试请求
	req := &ChatRequest{
		Messages: []Message{
			{Role: "system", Content: "You are a helpful assistant."},
			{Role: "user", Content: "Hello! Please introduce yourself briefly."},
		},
		MaxTokens:   100,
		Temperature: 0.7,
	}

	ctx := context.Background()

	// 演示阻塞式调用
	fmt.Println("\n🔄 Testing blocking chat...")
	resp, err := manager.ChatWithDefault(ctx, req)
	if err != nil {
		log.Printf("❌ Chat error: %v", err)
	} else {
		if resp.Error != nil {
			log.Printf("❌ API error: %s", resp.Error.Message)
		} else if len(resp.Choices) > 0 {
			fmt.Printf("✅ Response: %s\n", resp.Choices[0].Message.Content)
		}
	}

	// 演示流式调用
	fmt.Println("\n🌊 Testing streaming chat...")
	eventChan, err := manager.StreamChatWithDefault(ctx, req)
	if err != nil {
		log.Printf("❌ Stream chat error: %v", err)
		return
	}

	fmt.Print("🔤 Streaming response: ")
	for event := range eventChan {
		if event.Error != nil {
			log.Printf("\n❌ Stream error: %v", event.Error)
			break
		}

		if event.Data != nil {
			if event.Data.Error != nil {
				log.Printf("\n❌ API error: %s", event.Data.Error.Message)
				break
			}

			if len(event.Data.Choices) > 0 && event.Data.Choices[0].Delta != nil {
				fmt.Print(event.Data.Choices[0].Delta.Content)
			}
		}

		if event.Done {
			fmt.Println("\n✅ Stream completed")
			break
		}
	}

	// 清理资源
	if err := manager.Close(); err != nil {
		log.Printf("⚠️ Failed to close manager: %v", err)
	}

	fmt.Println("\n🎉 LLM demo completed!")
}
