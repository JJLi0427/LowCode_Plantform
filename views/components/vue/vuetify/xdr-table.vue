<template>
  <v-card v-if="items.length>0">
    <v-card-title>
      {{title}}
      <v-spacer></v-spacer>
      <v-text-field
        v-model="search"
        append-icon="mdi-magnify"
        :label="hint"
        single-line
        hide-details
      ></v-text-field>
    </v-card-title>
    <v-data-table
      :headers="headers"
      :items="items"
      :search="search"
    ></v-data-table>
  </v-card>
</template>

<script>
  export default {
    props:{ items: Array, title:{type: String, default: "数据表"}, hint: { type: String, default: "表数据检索"}, },
    data () {
      return {
        search: '',
        lists: [],
      }
    },
    watch:{
      lists(newlists) {
        this.items = newlists
      }
    },
    computed: {
      headers() {
        let h = []
        if (this.items.length > 0) {
          let first = true
          for (let key in this.items[0]) {
            let it = {}
            it.text = key
            it.value = key
            if (first){
              first = false
              it.align= 'start'
            }
            h.push(it)
          }
        }
        return h
      }
    },
  }
/**
 * 
 * Props: {item: Array}
  [
   { "name": "Frozen Yogurt",      "calories": 159, "fat": 6.0, "carbs": 24, "protein": 4.0, "iron": "1%" },
   { "name": "Ice cream sandwich", "calories": 159, "fat": 6.0, "carbs": 24, "protein": 4.0, "iron": "1%" },
   { "name": "Eclair",             "calories": 262, "fat": 1.6, "carbs": 23, "protein": 6.0, "iron": "7%" }
  ]
 */

</script>