<template>
  <v-container v-if="items.length>0">
    <v-row><v-col cols="12">
  <v-card
    class="md-auto"
  >
    <v-list three-line>
      <v-list-item-group
        v-model="selected"
        active-class="pink--text"
        multiple
      >
 
      <template v-for="(item, index) in items">
        <v-subheader
          v-if="item.header"
          :key="item.header"
          v-text="item.header"
        ></v-subheader>

        <v-divider
          v-else-if="item.divider"
          :key="index"
          :inset="item.inset"
        ></v-divider>

        <v-list-item
          v-else
          :key="'i'+index"
          @click="onclick(index)"
        >

        <template v-if="item?.thumbnail" >
            <v-img max-height="300" max-width="150" min-width="100" :src="item.thumbnail"></v-img>
        </template>
        <template v-else-if="item?.avatar" >
          <v-list-item-avatar>
            <v-img :src="item.avatar"></v-img>
          </v-list-item-avatar>
        </template>

          <v-list-item-content>
            <v-list-item-title v-html="item.title"></v-list-item-title>
            <v-list-item-subtitle v-html="item.subtitle"></v-list-item-subtitle>
          </v-list-item-content>
        </v-list-item>
      </template>
      </v-list-item-group>
    </v-list>
  </v-card>
</v-col></v-row></v-container>
</template>

<script>
  export default {
    props:{ items: Array },
    data: () => ({
      selected: [],
      lists: [],
    }),

    watch:{
      lists(newlists) {
        this.items = newlists
      }
    },

    methods: {
      onclick(index) {
        //TODO: ACTION on click

        console.log(index)
      }
    }

  }

/*
    define of props:{ items: Array } :
      [
        { header: 'Today' },
        {
          link | url : '',
          avatar: 'https://cdn.vuetifyjs.com/images/lists/1.jpg',
          title: 'Brunch this weekend?',
          subtitle: `<span class="text--primary">Ali Connors</span> &mdash; I'll be in your neighborhood doing errands this weekend. Do you want to hang out?`,
        },
        {
          link | url : '',
          thumbnail: 'https://cdn.vuetifyjs.com/images/lists/1.jpg',
          title: 'Brunch this weekend?',
          subtitle: `<span class="text--primary">Ali Connors</span> &mdash; I'll be in your neighborhood doing errands this weekend. Do you want to hang out?`,
        },
        { divider: true, inset: true }
      ]
*/

</script>
