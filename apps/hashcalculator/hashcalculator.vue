<template>
<div id="hashcontainer">
  <div class="hash-calculator">
    <h1>HASH 计算工具</h1>
    <select v-model="selectedAlgorithm">
      <option value="md5">MD5</option>
      <option value="sha1">SHA1</option>
      <option value="sha256">SHA256</option>
      <option value="sha512">SHA512</option>
    </select>
    <textarea v-model="inputData" placeholder="在此输入数据..."></textarea>
    <input type="file" @change="handleFileUpload" />
    <button @click="calculateHash">计算 HASH</button>
    <div v-if="hashResult">
      <h2>计算结果:</h2>
      <p>{{ hashResult }}</p>
    </div>
  </div>
</div>
</template>

<script>
export default {
  data() {
    return {
      selectedAlgorithm: 'sha-256', // 默认使用 SHA-256
      inputData: '',
      hashResult: ''
    };
  },
  methods: {
    async calculateHash() {
      const encoder = new TextEncoder();
      const data = encoder.encode(this.inputData);
      let algorithm;

      switch (this.selectedAlgorithm) {
        case 'md5':
          algorithm = 'MD5';
          break;
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

      const hashBuffer = await crypto.subtle.digest(algorithm, data);
      const hashArray = Array.from(new Uint8Array(hashBuffer));
      this.hashResult = hashArray.map(b => b.toString(16).padStart(2, '0')).join('');
    },
    handleFileUpload(event) {
      const file = event.target.files[0];
      if (file) {
        const reader = new FileReader();
        reader.onload = (e) => {
          this.inputData = e.target.result;
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
  max-width: 600px;
  margin: auto;
  padding: 20px;
  border: 1px solid #ccc;
  border-radius: 5px;
  background-color: #f9f9f9;
}
textarea {
  width: 100%;
  height: 100px;
  margin-bottom: 10px;
}
</style>