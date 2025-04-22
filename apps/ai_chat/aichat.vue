<template>
  <div id="chatcontainer">
    <div class="chat-interface">
      <h1>ChatBot</h1>
      <div class="chat-history">
        <div v-for="(message, index) in messages" :key="index" class="message" :class="message.role">
          <div>{{ message.role === 'user' ? '你：' : 'AI: ' }}</div>
          <div class="message-content">{{ message.content }}</div>
        </div>
      </div>
      
      <div class="input-container">
        <textarea 
          v-model="userInput" 
          placeholder="Input..." 
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
      apiKey: '',
      systemPrompt: "我建立了一个低代码配置文件框架, 可以通过一些简单的配置生成前端页面，你要学会下面这些内容, 然后在用户提问时做出相应的回答, 下面是配置文件的样例和一些注释: "
    };
  },
  mounted() {
    fetch('./sample.txt')
      .then(response => {
        if (!response.ok) {
          throw new Error(`无法加载配置文件: ${response.status}`);
        }
        return response.text();
      })
      .then(text => {
        this.systemPrompt += "\n" + text;
      })
      .catch(error => {
        console.error('加载sample.txt出错:', error);
      });
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
            model: 'deepseek-reasoner',
            messages: [
              { role: 'system', content: this.systemPrompt },
              ...this.messages
            ],
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
  background-color: rgb(0, 128, 255);
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