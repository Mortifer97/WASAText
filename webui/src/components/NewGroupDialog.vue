<template>
    <!-- Dialog per creare un nuovo gruppo. -->
    <!-- Sintassi base, commenti in italiano. -->
    <div
      class="modal fade"
      tabindex="-1"
      role="dialog"
      ref="modal"
      aria-labelledby="newGroupModalLabel"
      aria-hidden="true"
    >
      <div class="modal-dialog" role="document">
        <div class="modal-content">
          <div class="modal-header">
            <h5 class="modal-title" id="newGroupModalLabel">Crea Nuovo Gruppo</h5>
            <button type="button" class="btn-close" @click="close"></button>
          </div>
          <div class="modal-body">
            <div class="mb-3">
              <label for="groupName" class="form-label">Nome Gruppo</label>
              <input
                type="text"
                id="groupName"
                class="form-control"
                v-model="groupName"
                placeholder="Inserisci nome gruppo"
              />
            </div>
            <div class="mb-3">
              <label class="form-label">Seleziona Utenti</label>
              <div class="user-list">
                <div
                  v-for="user in searchResults"
                  :key="user.userId"
                  class="user-item d-flex justify-content-between align-items-center"
                  @click="toggleUserSelection(user)"
                >
                <div class="d-flex align-items-center">
                  <div class="user-photo me-3">
                  <img v-if="user.photo" :src="`data:image/png;base64,${user.photo}`" alt="Foto Utente" />
                  <i v-else class="bi bi-person-circle" style="font-size: 30px;"></i>
                  </div>
                  <span>{{ user.name }}</span>
                  <span v-if="user.userId == userId" class="badge">TU</span>
                </div>
                <input
                  type="checkbox"
                  :value="user.userId"
                  :checked="isSelected(user) || user.userId == userId"
                  @click.stop="toggleUserSelection(user)"
                />
                </div>
              </div>
            </div>
          </div>
          <div class="modal-footer">
            <button type="button" class="btn btn-secondary" @click="close">Chiudi</button>
            <button
              type="button"
              class="btn btn-primary"
              @click="createGroup"
              :disabled="!groupName || selectedUsers.length === 0"
            >
              Crea Gruppo
            </button>
          </div>
        </div>
      </div>
    </div>
  </template>
  
<script>
  import { searchUsers } from "@/services/axios";
  export default {
    data() {
      return {
        groupName: "",
        selectedUsers: [],
        searchResults: [],
      };
    },
    methods: {
      open() {
        const modal = new bootstrap.Modal(this.$refs.modal);
        modal.show();
        this.reset();

        // Ottieni userId da localStorage
        const userId = localStorage.getItem("userId");
        this.userId = userId;
        this.search();
      },
      close() {
        const modal = bootstrap.Modal.getInstance(this.$refs.modal);
        modal.hide();
        this.reset();
      },
      reset() {
        this.groupName = "";
        this.selectedUsers = [];
        this.searchResults = [];
      },
      createGroup() {
        if (this.groupName && this.selectedUsers.length > 0) {
          this.$emit("create-group", {
            name: this.groupName,
            users: this.selectedUsers,
          });
          this.close();
        }
      },
      async search() {
        try {
          console.log("Ricerca utenti");
          const results = await searchUsers(this.userId, "");
          this.searchResults = results;
        } catch (error) {
          console.error("Errore durante la ricerca degli utenti:", error);
        }
      },
      isSelected(user) {
        return this.selectedUsers.some(selectedUser => selectedUser.userId === user.userId);
      },
      toggleUserSelection(user) {
        if (user.userId == this.userId) {
            alert("Non puoi deselezionare te stesso. Sei già nel gruppo.");
          return;
        }
        const index = this.selectedUsers.findIndex(selectedUser => selectedUser.userId === user.userId);
        if (index === -1) {
          // Aggiungi utente se non selezionato
          this.selectedUsers.push(user);
        } else {
          // Rimuovi utente se già selezionato
          this.selectedUsers.splice(index, 1);
        }
      },
    },
  };
  </script>
  
<style scoped>
@import '../assets/style.css';
.modal-content, .modal-body, .modal-header, .modal-footer {
  background: var(--color-bg-secondary) !important;
  color: var(--color-fg) !important;
  border-radius: 8px;
  border: 1px solid var(--color-border);
}
  .modal-content {
    border-radius: 0.5rem;
  }
  .user-list {
    max-height: 200px;
    overflow-y: auto;
  }
  .user-item {
    padding: 10px;
    border-bottom: 1px solid #ddd;
  }
  .user-item:last-child {
    border-bottom: none;
  }
  .user-photo img {
    width: 100px;
    height: 100px;
    object-fit: cover;
    border-radius: 50%;
  }
  .badge {
    background-color: #17a2b8;
    color: white;
    font-size: 12px;
    padding: 3px 7px;
    border-radius: 3px;
    margin-left: 10px;
  }
</style>
