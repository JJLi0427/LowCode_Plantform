<template>
      <div id="mainbox">
      <h1 style="text-align:center;">json，xml格式化</h1>
      <v-divider></v-divider>
      <!-- <v-form style="border:1px solid gray; border-redius:1px;"> -->
      <v-container fluid>
            <v-row style="flex-wrap: nowrap;" >
                  <v-col :style="`flex-grow: ${leftflexgrow};`" >
                        
                        <v-toolbar dense rounded elevation="3">
                               <v-tooltip bottom>
                                    <template v-slot:activator="{ on, attrs }">
                                          <v-btn 
                                                icon 
                                                @click="boxgrowshrink(-2)"
                                                v-bind="attrs"
                                                v-on="on">
                                          <v-icon>mdi-arrow-collapse-horizontal</v-icon>
                                          </v-btn>
                                          
                                    </template>
                                    <span>缩小单元宽度</span>
                              </v-tooltip>
                               <v-tooltip bottom>
                                    <template v-slot:activator="{ on, attrs }">
                                          <v-btn 
                                                icon 
                                                @click="boxgrowshrink(2)"
                                                v-bind="attrs"
                                                v-on="on">
                                          <v-icon>mdi-arrow-expand-horizontal</v-icon>
                                          </v-btn>
                                          
                                    </template>
                                    <span>扩大单元宽度</span>
                              </v-tooltip>
                                    <!-- <v-app-bar-nav-icon></v-app-bar-nav-icon> -->
                                    <v-btn>
                                          <v-select
                                          style="width:100px;"
                                          v-model="select"
                                          :items="items"
                                          item-text="texttype"
                                          item-value="value"
                                          label="Select"
                                          @change="selected"
                                          hide-details=true
                                          return-object
                                          single-line
                                          ></v-select>
                                    </v-btn>
                              </v-toolbar>
                      
                        <v-divider></v-divider>
                        <v-textarea
                              filled
                              clearable
                              clear-icon="mdi-close-circle"
                              name="input-textarea"
                              label="输入内容"
                              rows=20
                              dense
                              v-model="messagestring"
                              counter
                              @change="show"

                        ></v-textarea>
                  </v-col>
                        
                  <v-col style="flex-grow: 10;">
                        <v-toolbar dense rounded elevation="3">
                              <!-- <v-spacer></v-spacer> -->
                               <v-tooltip bottom>
                                    <template v-slot:activator="{ on, attrs }">
                                          <v-btn 
                                                icon 
                                                @click="changeFoldClosed($event)"
                                                v-bind="attrs"
                                                v-on="on">
                                          <v-icon>mdi-unfold-less-horizontal</v-icon>
                                          </v-btn>
                                          
                                    </template>
                                    <span>折叠所有</span>
                              </v-tooltip>
                              <v-tooltip bottom>
                                    <template v-slot:activator="{ on, attrs }">
                
                                          <v-slider
                                          v-model="slidervalue"
                                          class="align-center"
                                          :max="slidermax"
                                          :min="slidermin"
                                          hide-details
                                          thumb-label
                                          @change="foldToLevel"
                                          v-bind="attrs"
                                          v-on="on">
                                                <template v-slot:append>
                                                      <v-text-field
                                                      v-model="slidervalue"
                                                      class="mt-0 pt-0"
                                                      hide-details
                                                      single-line
                                                      type="number"
                                                      style="width: 60px"
                                                      @change="foldToLevel"
                                                      ></v-text-field>
                                                </template>
                                          </v-slider>
                                    </template>
                                    <span>展示深度</span>
                              </v-tooltip>
                              <v-tooltip bottom>
                                    <template v-slot:activator="{ on, attrs }">
                                          <v-btn 
                                                icon 
                                                @click="changeFoldOpen($event)"
                                                v-bind="attrs"
                                                v-on="on">
                                                <v-icon>mdi-unfold-more-horizontal</v-icon>
                                          </v-btn>
                                    </template>
                                    <span>展开所有</span>
                              </v-tooltip>
                              
                              
                              <v-tooltip bottom>
                                    <template v-slot:activator="{ on, attrs }">
                                          <v-btn 
                                                icon 
                                                @click="copyToClipboard"
                                                v-bind="attrs"
                                                v-on="on">
                                          <v-icon>mdi-content-copy</v-icon>
                                          </v-btn>
                                    </template>
                                    <span>复制</span>
                              </v-tooltip>
                        </v-toolbar>
                         <v-snackbar
                                    v-model="snackbarcopysuccess"
                                    :timeout="timeout"
                                    :top=true
                                    color='success'
                                    shaped
                                    elevation="40"
                              >
                                    {{ textsuccess }}
                         </v-snackbar>
                        <v-snackbar
                                    v-model="snackbarcopyfail"
                                    :timeout="timeout"
                                    :top=true
                                    color="fail"
                                    shaped
                                    elevation="40"
                              >
                                    {{ textfail }}
                         </v-snackbar>
                         <v-divider></v-divider>
                         <div id="editor" style="border:1px solid gray;"></div>
                  </v-col>
                  
            </v-row>
            
      </v-container>

      
      </div>
</template>
<style>
div.ace_content{
  height: 100% !important;
  width: 100% !important;
}
div#maincontainer {
  width: 98vw;
  height: 96vh;
}
div#editor{
  height: 89% !important;
}
#mainbox {
  width: 100%;
  max-height: 96vh;
}

@media (min-width: 1364px){
.container {
  max-width: 96vw;
}
}

@media (min-width: 1864px){
.container {
  max-width: 1700px;
}
}

</style>


<script>
export default {
      name: 'JsonFormat',
      data() {
      return {
            messageobj: {} ,
            messagestring: "",
            acevalue:"",
            editor:null,
            select: { texttype: 'json', value: 'json' },
            items: [
                  { texttype: 'json', value: 'json' },
                  { texttype: 'xml', value: 'xml' },
                  { texttype: 'text', value: 'text' },
                  { texttype: 'auto', value: 'auto' }
            ],
            leftflexgrow:10,
            slidermin: 1,
            slidermax: 10,
            slidervalue: 1,
            autoMode:false,
            snackbarcopysuccess: false,
            snackbarcopyfail: false,
            textsuccess: '复制成功',
            textfail:'复制失败',
            timeout: 1000,
            timer:null,
            }
      },
      watch:{
        messagestring(newval, oldval) {
            newvalLength = newval? newval.length:0
            oldvalLength = oldval? oldval.length:0

          if (Math.abs(newvalLength-oldvalLength) > 0) {
            if (this.timer != null) {
              clearTimeout(this.timer)
              this.timer = null
            }
            this.timer = setTimeout(()=>{this.show(); this.timer = null;}, 700)
          }
        }
      },
      methods: {
            boxgrowshrink:function(delta){
              this.leftflexgrow += delta
              if (this.leftflexgrow < 1){
                  this.leftflexgrow = 1;
              }
              if (this.leftflexgrow > 40){
                  this.leftflexgrow = 40;
              }
            },
            selected: function () {
            // console.log(this.select.texttype)
            switch (this.select.texttype) {
                  case 'json': {
                        var JsonMode = ace.require("ace/mode/json").Mode;
                        this.editor.getSession().setMode(new JsonMode());
                        break
                  }

                  case 'text': {
                        var JavaScriptMode = ace.require("ace/mode/text").Mode;
                        this.editor.getSession().setMode("ace/mode/text");
                        break
                  }

                  case 'xml': {
                        var xmlMode = ace.require("ace/mode/xml").Mode;
                        this.editor.getSession().setMode(new xmlMode());
                        break
                  }


            }
            this.show()
            
            },
            show: function () {
            // console.log("转换")
            // console.log(this.messagestring)
            // console.log(this.select.texttype)
            if (this.select.texttype == 'auto')
            {            
                  strtype = check_string_type(this.messagestring).toLowerCase()
                  // console.log('ace/mode/'+strtype.toLowerCase())
                  var xmlMode = ace.require('ace/mode/'+strtype).Mode;
                  this.editor.getSession().setMode(new xmlMode());
                  this.formatStr(strtype, this.messagestring)
                  
            }else{

                  this.formatStr(this.select.texttype, this.messagestring)
            }
            
            
            
            },
            formatStr:function(strType,messagestr)
            {
                  if (messagestr == null ||messagestr.length == 0) {
                      this.editor.session.setValue("");
                      return
                  }
                  switch (strType) {
                        case 'json': {
                            try {
                              JSON.parse(messagestr)
                            } catch (error) {
                              this.editor.session.setValue(`${error.name} ${error.message}`)
                              break
                            }
                            this.acevalue = string_to_json_wrap(messagestr)
                            this.editor.session.setValue(this.acevalue)
                            break
                        }

                        case 'text': {
                              this.acevalue = messagestr
                              this.editor.session.setValue(this.acevalue)
                              break
                        }

                        case 'xml': {
                              this.acevalue = string_to_xml_wrap(messagestr)
                              this.editor.session.setValue(this.acevalue)
                              break
                        }
                  }
            },
            copyToClipboard:function(){
                  if (navigator && navigator.clipboard && navigator.clipboard.writeText)
                  {
                         if(navigator.clipboard.writeText(this.acevalue)) 
                        {
                              this.snackbarcopysuccess=true;
                        }
                        else{
                              this.snackbarcopyfail=true;
                        }
                        
                  }
                  
                       
                  
            },
            changeFoldClosed:function(event){

                  this.editor.getSession().foldAll()
                  // this.editor.getSession().toggleFold(true)

            },
            changeFoldOpen:function(event){
                  this.editor.getSession().unfold()
            },
            foldToLevel:function(event){
                  this.editor.getSession().foldToLevel(this.slidervalue)
            }
      },
      
      mounted(){
            this.editor = ace.edit("editor");
            //var JsonMode = ace.require("ace/mode/json").Mode;
            // ace.config.set('basePath', '../src-min')
            this.editor.getSession().setMode("ace/mode/json");
            editorElement = document.getElementById('editor');
            editorElement.style.fontSize='1rem';


      },

      // watch: {
      //       // 每当 question 发生变化时，该函数将会执行
      //       messagestring(newQuestion, oldQuestion) {
      //             console.log(newQuestion)
      //             try{
      //                   this.messageobj = JSON.parse(newQuestion)
      //             }
      //             catch (error){
      //                   alert("输入json格式不正确");
      //             }
                  
      //       },
           
      // },
}
</script>