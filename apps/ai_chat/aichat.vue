<template>
  <div id="chatcontainer">
    <div class="chat-interface">
      <h1>ChatGPT 聊天</h1>
      <div class="chat-history">
        <div v-for="(message, index) in messages" :key="index" class="message" :class="message.role">
          <strong>{{ message.role === 'user' ? '你：' : 'ChatGPT: ' }}</strong>
          <div class="message-content">{{ message.content }}</div>
        </div>
        <div v-if="isLoading" class="message assistant">
          <strong>ChatGPT: </strong>
          <div class="message-content">{{ currentResponse }}<span class="cursor"></span></div>
        </div>
      </div>
      
      <div class="input-container">
        <textarea 
          v-model="userInput" 
          placeholder="输入你的问题..." 
          @keyup.enter.ctrl="sendMessage"
        ></textarea>
        <div class="button-container">
          <button @click="sendMessage" :disabled="isLoading || !userInput.trim()">发送</button>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
export default {
  data() {
    return {
      userInput: '',
      messages: [],
      currentResponse: '',
      isLoading: false,
      apiKey: ''
    };
  },
  methods: {
    async sendMessage() {
      if (!this.userInput.trim() || this.isLoading) return;
      
      // 添加用户消息到聊天记录
      const userMessage = { role: 'user', content: this.userInput.trim() };
      this.messages.push(userMessage);
      
      // 清空输入框和准备接收响应
      const userPrompt = this.userInput;
      this.userInput = '';
      this.isLoading = true;
      this.currentResponse = '';
      
      try {
        // 准备消息历史
        const messageHistory = this.messages.map(msg => ({
          role: msg.role === 'assistant' ? 'assistant' : 'user',
          content: msg.content
        }));
        
        // 发送请求到OpenAI API
        const response = await fetch('https://api.openai.com/v1/chat/completions', {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json',
            'Authorization': `Bearer ${this.apiKey}`
          },
          body: JSON.stringify({
            model: 'gpt-4',
            messages: messageHistory,
            stream: true // 开启流式响应
          })
        });
        
        // 处理流式响应
        const reader = response.body.getReader();
        const decoder = new TextDecoder('utf-8');
        
        while (true) {
          const { done, value } = await reader.read();
          if (done) break;
          
          // 解析并累加响应内容
          const chunk = decoder.decode(value);
          const lines = chunk.split('\n').filter(line => line.trim() !== '');
          
          for (const line of lines) {
            if (line.includes('[DONE]') || !line.startsWith('data:')) continue;
            
            try {
              const jsonData = JSON.parse(line.substring(5));
              const content = jsonData.choices[0]?.delta?.content;
              if (content) {
                this.currentResponse += content;
              }
            } catch (e) {
              console.error('解析响应失败:', e);
            }
          }
        }
        
        // 将完整响应添加到消息历史
        this.messages.push({ role: 'assistant', content: this.currentResponse });
      } catch (error) {
        console.error('API请求失败:', error);
        this.messages.push({ role: 'assistant', content: '抱歉，请求失败，请稍后再试。' });
      } finally {
        this.isLoading = false;
        this.currentResponse = '';
      }
    }
  }
};
</script>

<style scoped>
div#chatcontainer {
  font-size: large;
  width: 100%;
}

div.chat-interface {
  margin: auto;
  padding: 20px;
  border: 5px solid transparent;
  border-radius: 5px;
  background-color: #f9f9f9;
  display: flex;
  flex-direction: column;
  height: 80vh;
}

.chat-history {
  flex-grow: 1;
  overflow-y: auto;
  margin-bottom: 15px;
  padding: 10px;
  background-color: #ffffff;
  border-radius: 5px;
  border: 1px solid #e0e0e0;
}

.message {
  margin-bottom: 15px;
  padding: 10px;
  border-radius: 5px;
}

.message.user {
  background-color: #e6f7ff;
  align-self: flex-end;
}

.message.assistant {
  background-color: #f0f0f0;
  align-self: flex-start;
}

.message-content {
  margin-top: 5px;
  white-space: pre-wrap;
}

.cursor {
  display: inline-block;
  width: 8px;
  height: 16px;
  background-color: #000;
  animation: blink 1s infinite;
  vertical-align: middle;
}

@keyframes blink {
  0%, 100% { opacity: 1; }
  50% { opacity: 0; }
}

.input-container {
  display: flex;
  flex-direction: column;
}

textarea {
  width: 100%;
  height: 80px;
  margin-bottom: 10px;
  padding: 10px;
  border-radius: 5px;
  border: 1px solid #d9d9d9;
  resize: none;
}

.button-container {
  display: flex;
  justify-content: flex-end;
}

button {
  border-radius: 5px;
  background-color: gray;
  color: white;
  border: none;
  padding: 10px 20px;
  cursor: pointer;
}

button:disabled {
  background-color: #cccccc;
  cursor: not-allowed;
}
</style>
