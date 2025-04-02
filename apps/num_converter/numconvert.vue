<template>
<div id="numcontainer">
  <h1> 进制转换器 </h1>
  <div id="req">

  <div class="label"> 选择进制数，并输入: </div>
  <select v-model="selectedval">
    <option value="2">二进制</option>
    <option value="8">八进制</option>
    <option value="10" selected="selected">十进制</option>
    <option value="16">十六进制</option>
    <option value="32">三十二进制</option>
  </select>
  <input type="text" v-model="inputdata" />
  </div>

  <div id="results">
  <div class="label"> 转换结果: </div>
  <div><span class="title">二进制</span><span class="num">{{get2}}</span></div>
  <div><span class="title">八进制</span><span class="num">{{get8}}</span></div>
  <div><span class="title">十进制</span><span class="num">{{get10}}</span></div>
  <div><span class="title">十六进制</span><span class="num">{{get16}}</span></div>
  <div><span class="title">三十二进制</span><span class="num">{{get32}}</span></div>
  </div>
</div>
</template>

<style scoped>
body{
  background-color: rgb(245, 245, 245);
}
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

div.label{
  font-weight: 600;
}

#req, #results{
  width: 100%;
  margin-bottom: 1rem;
}

div#numcontainer {
  font-size: large;
  width: 100%;
}

div#numcontainer h1 {
  text-align: center;
}

div#req select:focus, div#req input:focus {
  border: 1px solid rgba(87, 131, 245, 0.5);
  box-shadow: 0px 0px 5px rgba(87, 131, 245, 0.5);
}

div#req select, div#req input {
  outline:none;
  font-size: large;
  height: 2rem;
  border: 1px solid rgba(175, 174, 176, 0.6);
  background-color: white  ;
  border-radius: 0.5rem;
}

#results div:nth-child(2n){
  border-radius: 5px;
  background-color: rgb(214, 214, 214);
}

#results div span.title {
  display: inline-block;
  width: 8rem;
}


#results div span.num {
  font-weight: bolder;
}

</style>

<script>
export default {
  name:'numconvert',
  data(){
    return {
      selectedval:'10',
      inputdata:'',
      inputval: 0,
    }
  },

  watch:{
    selectedval: function(snew, old) {
      this.inputdata='';
      //this.inputval =  parseInt(this.inputdata, parseInt(this.selectedval))
    },

    inputdata: function(snew, old) {
      if (snew.length == 0) {
        this.inputval = 0;
        return;
      }

      if (this.selectedval == '10') {
        if (!/^[0-9\-]+$/.test(snew)) {
          alert("请输入合法的十进制数");
          this.inputdata = '';
          return;
        }
      }
      else if (this.selectedval == '16') {
        if (!/^[A-Fa-f0-9]+$/.test(snew)){
          alert("请输入合法的十六进制数");
          this.inputdata = '';
          return;
        }
      }
      else if (this.selectedval == '8') {
        if (!/^[0-7]+$/.test(snew)){
          alert("请输入合法的八进制数");
          this.inputdata = '';
          return;
        }
      }
      else if (this.selectedval == '2') {
        if (!/^(0[bB])?[0-1]+$/.test(snew)){
          alert("请输入合法的二进制数");
          this.inputdata = '';
          return;
        }
      }

      this.inputval =  parseInt(snew, parseInt(this.selectedval));
    }
  },

  computed: {
    get2() {
      return this.inputval.toString(2)
    },

    get8() {
      return this.inputval.toString(8)
    },

    get10() {
      return this.inputval.toString(10)
    },

    get16() {
      return this.inputval.toString(16)
    },

    get32() {
      return this.inputval.toString(32)
    }

  }

}
  
</script>