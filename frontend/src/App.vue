<template>
  <div id="app" class="container">
    <div class="row">
      <div class="col-md-6 offset-md-3 py-5">
        <h1>Send quick-note</h1>

        <form v-on:submit.prevent="sendMail()">
          <div class="form-group">
            <input v-model="sendText" type="text" id="website-input" placeholder="Enter note" class="form-control">
          </div>

          <div class="form-group">
            <button class="btn btn-primary">Send!</button>
          </div>
          <span v-html="serverResponse"></span>
        </form>
      </div>
    </div>
  </div>
</template>

<script>
import axios from 'axios';
import { Config } from './config';


export default {
  name: 'App',

  data() {
    return {
      sendText: '',
      // serverResponse: `${Config.apiHostname}`,
      serverResponse: "",
    }
  },

  methods: {
    sendMail() {

      axios({
        url: `${Config.apiHostname}`,
        method: "POST",
        output: 'json',
        data: {
          todo: this.sendText}
      })
        .then((response) => {
          this.serverResponse = response.data.msg;
        })
        .catch((error) => {
          this.serverResponse = `Error: ${error}<br/>${error.response.data.msg}`
        })

    }
  }
}

// HOW TO USER THIS FUNCTION?
// function getHealth() {   

//   var x = axios.get('http://localhost:3000/health')
//   .then((response) => {
//     x = response.data.version
//   });

//   return x
// } 

</script>
