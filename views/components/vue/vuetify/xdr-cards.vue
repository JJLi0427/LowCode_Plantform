<template>
<v-container> <v-row>
  <v-card
    v-for="(item, idx) in lists" :key="idx"
    class="mx-auto my-12"
    max-width="374"
    :href="item?.target ? item.target:''"
    target="_blank"
  >
    <v-img
      v-if="item?.image"
      height="250"
      :src="item.image"
    ></v-img>
    <v-img
      v-if="item?.thumbnail"
      height="150"
      :src="item.thumbnail"
    ></v-img>

    <v-list-item v-if="item?.avater" >
    <v-list-item-avatar>
     <v-img 
      :src="item.avater"
     ></v-img>
    </v-list-item-avatar>
            <v-list-item-title v-if="item?.title" v-html="item.title"></v-list-item-title>
            <v-list-item-title v-if="item?.head"  v-html="item.head"></v-list-item-title>
    </v-list-item>

    <v-card-title v-if="!item?.avater && item?.title" v-html="item.title"></v-card-title>
    <v-card-title v-if="!item?.avater && item?.head" v-html="item.head"></v-card-title>

    <v-card-text v-if="item?.rate">
      <v-row
        align="center"
        class="mx-0"
      >
        <v-rating
          :value="item.rate?.score ? item.rate.score : item.rate"
          color="amber"
          dense
          half-increments
          readonly
          size="14"
        ></v-rating>

        <div class="grey--text ms-4">
	  {{item.rate?.score ? item.rate.score : item.rate}}
        </div>
      </v-row>

      <div class="my-4 text-subtitle-1" v-if="item?.subtitle">
	{{item.subtitle}}
      </div>

      <div v-if="item?.summary" >{{cvtsummary(item.summary)}}</div>
      <div v-if="item?.text" >{{item.text}}</div>
      <div v-if="item?.detail" >{{item.detail}}</div>
    </v-card-text>

    <v-card-subtitle v-if="!item?.rate && item?.subtitle" v-html="item.subtitle"></v-card-subtitle>
    <v-card-text v-if="!item?.rate">
      <div v-if="item?.summary" >{{cvtsummary(item.summary)}}</div>
      <div v-if="item?.text" >{{item.text}}</div>
      <div v-if="item?.detail" >{{item.detail}}</div>
    </v-card-text>


    <template v-if="item?.url || item?.link">
    <v-divider class="mx-4"></v-divider><v-card-actions >
      <v-btn
        color="deep-purple lighten-2"
        text
        @click="onclick(idx)"
      >
      Learn More
      </v-btn>
    </v-card-actions></template>
  </v-card>
</v-row> </v-container>
</template>

<script>
  export default {
    props:{ items: Array },
    data: () => ({
      lists:[],
    }),

    watch:{
	items(newitems) {
		this.lists = newitems
	}
    },

    methods: {
      cvtsummary(txt) {
        if (txt.length > 100) {
          return txt.substring(0, 100)+" ..."
        }
        return txt
      },
      onclick(index) {
        let link = this.lists[index]?.link ?  this.lists[index].link : this.lists[index].url
        if (link.length > 0) {
          window.open(link, '_blank',)
        }
      }
    },
  }

/* 
    props:{ items: Array }
[{
	title | head | subtitle :
	text | detail | summary 
	price| amount:
	rate: {score: 5.0, total: 100 } | rate: 1.5
	image | avater | thumbnail
	target | link | url
}]

*/
</script>