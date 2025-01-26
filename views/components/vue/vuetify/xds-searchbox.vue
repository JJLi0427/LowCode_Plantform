<template>
  <v-form>
    <v-container>
      <v-row v-if="hastitle"><v-col cols="12"><h1 id="title">{{title}}</h1></v-col></v-row>
      <v-row>
        <v-col cols="12">
          <v-text-field
            solo
            clearable
            :disabled="isloading"
            :loading="isloading"
            :label="hint"
            v-model="message"
            :append-outer-icon="message ? 'mdi-send' : ''"
            prepend-inner-icon="mdi-magnify"
            clear-icon="mdi-close-circle"
            type="text"
            @click:append-outer="sendMessage"
            @click:clear="clearMessage"
            @keydown.enter.prevent="sendMessage"
          ></v-text-field>
        </v-col>
      </v-row>
    </v-container>
    <slot :results="Results" ></slot>
  </v-form>
</template>


<script>
  export default {
    props:{
      title: String,
      posturl: String,
      hint: String,
      initFetchArgument: String,
    },
    data: () => ({

      isloading: false,
      message: '',
      Results: [],
    }),

    computed: {
      hastitle () { return this.title.length > 0 }
    },

    mounted() {
      if (this.initFetchArgument.length > 0) {
          this.send(this.initFetchArgument)
      }
    },

    methods: {
      sendMessage () {
        this.send(this.message)
      },
      send(message) {
        if (this.posturl.length > 0 && message.length > 0) {
          let self = this
          self.isloading = true
          var payload = new FormData();payload.append('search', message)
          AsyncPostJsonData(this.posturl, payload, (res)=>{
            if (self.$scopedSlots.default) {
              if (res instanceof Array) {
                self.Results = res
              } else if (res?.Results && res.Results instanceof Array) {
                self.Results = res.Results
              } else {
                for (let key in res) {
                  if (res[key] instanceof Array) {
                    self.Results = res[key]
                    break
                  }
                }
              }
            }
            else if (typeof dispatchData == 'function') {
              dispatchData(res)
            }
            else {
              $emit('ondata', res)
            }
            self.isloading = false
          })
        }
        //this.clearMessage()
      },
      clearMessage () {
        this.message = ''
      },
    },
  }
</script>


<style scoped>
#title{
  text-align: center;
}
</style>