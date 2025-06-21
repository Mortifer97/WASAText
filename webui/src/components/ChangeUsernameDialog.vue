<template>
    <!-- Dialog per cambiare username. -->
    <!-- Sintassi base, commenti in italiano. -->
    <div v-if="isOpen" class="modal fade show d-block" tabindex="-1" role="dialog" aria-labelledby="changeUsernameLabel" aria-hidden="true">
      <div class="modal-dialog" role="document">
        <div class="modal-content">
          <div class="modal-header">
            <h5 class="modal-title" id="changeUsernameLabel">Cambia Username</h5>
            <button type="button" class="btn-close" aria-label="Close" @click="closeDialog"></button>
          </div>
          <div class="modal-body">
            <form @submit.prevent="submitUsernameChange">
              <div class="mb-3">
                <label for="usernameInput" class="form-label">Nuovo Username</label>
                <input
                  type="text"
                  id="usernameInput"
                  class="form-control"
                  v-model="newUsername"
                  placeholder="Inserisci il nuovo username"
                  required
                />
              </div>
              <div class="d-flex justify-content-end">
                <button type="button" class="btn btn-secondary me-2" @click="closeDialog">Annulla</button>
                <button type="submit" class="btn btn-primary">Cambia</button>
              </div>
            </form>
          </div>
        </div>
      </div>
    </div>
    <div v-if="isOpen" class="modal-backdrop fade show"></div>
  </template>
  
  <script>
  export default {
    emits: ["username-changed"],
    data() {
      return {
        isOpen: false,
        newUsername: "",
      };
    },
    methods: {
      open() {
        this.newUsername = localStorage.getItem("username") || "";
        this.isOpen = true;
      },
      closeDialog() {
        this.isOpen = false;
        this.newUsername = "";
      },
      submitUsernameChange() {
        if (this.newUsername.trim() === "") {
          alert("L'username non può essere vuoto.");
          return;
        }
        this.$emit("username-changed", this.newUsername.trim());
        this.closeDialog();
      },
    },
  };
  </script>
  
  <style scoped>
  @import '../assets/style.css';
  .modal-backdrop {
    z-index: 1040;
  }
  
  .modal {
    z-index: 1050;
  }
  .modal-content, .modal-body, .modal-header, .modal-footer {
    background: var(--color-bg-secondary) !important;
    color: var(--color-fg) !important;
    border-radius: 8px;
    border: 1px solid var(--color-border);
  }
  </style>
