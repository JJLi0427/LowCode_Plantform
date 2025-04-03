<template>
  <div id="hashcontainer">
    <div class="hash-calculator">
      <h1>HASH 计算工具</h1>
      <select v-model="selectedAlgorithm">
        <option value="sha1">SHA1</option>
        <option value="sha256">SHA256</option>
        <option value="sha512">SHA512</option>
      </select>
      <textarea v-model="inputData" placeholder="在此输入数据..."></textarea>
      <div>
        <input type="file" @change="handleFileUpload" />
      </div>
      <div class="button-container">
        <button @click="calculateHash">计算 HASH</button>
      </div>
      <div v-if="hashResult">
        <h2>计算结果:</h2>
        <textarea v-model="hashResult" readonly></textarea> <!-- 新增的文本框 -->
      </div>
    </div>
  </div>
</template>

<script>
export default {
  data() {
    return {
      selectedAlgorithm: 'sha256', // 默认使用 SHA-256
      inputData: '',
      hashResult: ''
    };
  },
  methods: {
    async calculateHash(data) {
      const encoder = new TextEncoder();
      const encodedData = encoder.encode(data || this.inputData);
      let algorithm;

      switch (this.selectedAlgorithm) {
        case 'sha1':
          algorithm = 'SHA-1';
          break;
        case 'sha256':
          algorithm = 'SHA-256';
          break;
        case 'sha512':
          algorithm = 'SHA-512';
          break;
      }

      const hashBuffer = await crypto.subtle.digest(algorithm, encodedData);
      const hashArray = Array.from(new Uint8Array(hashBuffer));
      this.hashResult = hashArray.map(b => b.toString(16).padStart(2, '0')).join('');
    },
    handleFileUpload(event) {
      const file = event.target.files[0];
      if (file) {
        const reader = new FileReader();
        reader.onload = async (e) => {
          const fileContent = e.target.result;
          await this.calculateHash(fileContent);
        };
        reader.readAsText(file);
      }
    }
  }
};
</script>

<style scoped>
div#maincontainer {
  width: 98%;
  min-width: 600px;
}

div#vueinputelem {
  width: 70%;
  min-width: 600px;
  margin-left: auto;
  margin-right: auto;
  display: flex;
  flex-direction: column;
  justify-content: flex-start;
  align-items: center;
}

div#hashcontainer {
  font-size: large;
  width: 100%;
}

div.hash-calculator {
  margin: auto;
  padding: 20px;
  border: 5px solid transparent;
  border-radius: 5px;
  background-color: #f9f9f9;
}

textarea {
  width: 100%;
  height: 100px;
  margin-bottom: 10px;
}

.button-container {
  margin-bottom: 10px;
}

button {
  border-radius: 5px;
  background-color: gray;
  color: white;
  border: none;
  margin-top: 10px;
  padding: 10px 10px;
  cursor: pointer;
}

textarea {
  border-radius: 5px;
  border: 5px solid transparent;
  height: 100px;
  margin-top: 10px;
  margin-bottom: 10px;
}
</style>