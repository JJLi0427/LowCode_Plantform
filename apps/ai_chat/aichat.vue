<template>
  <div id="chatcontainer">
    <div class="chat-interface">
      <h1>大模型对话</h1>
      <div class="chat-history">
        <div v-for="(message, index) in messages" :key="index" class="message" :class="message.role">
          <div>{{ message.role === 'user' ? '你：' : 'AI: ' }}</div>
          <div class="message-content">{{ message.content }}</div>
        </div>
      </div>
      
      <div class="input-container">
        <textarea 
          v-model="userInput" 
          placeholder="输入你的问题..." 
          @keyup.enter.ctrl="sendMessage"
        ></textarea>
        <div class="button-container">
          <button @click="sendMessage" :disabled="isLoading || !userInput.trim()">
            {{ isLoading ? '请求中...' : '发送' }}
          </button>
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
      isLoading: false,
      apiBaseUrl: 'https://api.deepseek.com/v1',
      apiKey: ''
    };
  },
  methods: {
    async sendMessage() {
      if (!this.userInput.trim() || this.isLoading) return;

      const userMessage = { role: 'user', content: this.userInput.trim() };
      this.messages.push(userMessage);
      this.userInput = '';
      this.isLoading = true;

      try {
        const response = await fetch(`${this.apiBaseUrl}/chat/completions`, {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json',
            'Authorization': `Bearer ${this.apiKey}`
          },
          body: JSON.stringify({
            model: 'deepseek-chat',
            messages: this.messages,
            temperature: 0.7
          })
        });

        if (!response.ok) {
          throw new Error(`API请求失败: ${response.status} ${response.statusText}`);
        }

        const data = await response.json();
        const answer = data.choices[0]?.message?.content || '';
        this.messages.push({ role: 'assistant', content: answer });
      } catch (error) {
        console.error('请求错误:', error);
        this.messages.push({
          role: 'assistant',
          content: `请求失败: ${error.message}，请检查API KEY或稍后重试`
        });
      } finally {
        this.isLoading = false;
      }
    }
  }
};
</script>

<style scoped>
/* 样式保持不变 */
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