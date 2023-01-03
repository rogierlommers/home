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
          <span v-html="sendResponse"></span>
        </form>
      </div>
    </div>
  </div>
</template>

<script>
import axios from 'axios';
export default {
  name: 'App',

  data() {
    return {
      sendText: '',
      sendResponse: '',
    }
  },

  methods: {
    sendMail() {

      axios.post("http://quick-note.lommers.org/api/send", {
        text: this.sendText,
        output: 'json',
      })
        .then((response) => {
          this.sendResponse = response.data.status_text;
        })
        .catch((error) => {
          this.sendResponse = `Error: ${error}`
        })

    }
  }
}
</script>
