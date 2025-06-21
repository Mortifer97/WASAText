<template>
    <!-- Dialog per aggiungere membri a un gruppo. -->
    <!-- Sintassi base, commenti in italiano. -->
    <div
      class="modal fade"
      tabindex="-1"
      role="dialog"
      ref="modal"
      aria-labelledby="addMemberModalLabel"
      aria-hidden="true"
    >
      <div class="modal-dialog" role="document">
        <div class="modal-content">
          <div class="modal-header">
            <h5 class="modal-title" id="addMemberModalLabel">Aggiungi Membri</h5>
            <button type="button" class="btn-close" @click="close"></button>
          </div>
          <div class="modal-body">
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
                      <img v-if="user.photo" :src="`data:image/png;base64,${user.photo}`" alt="User Photo" />
                      <i v-else class="bi bi-person-circle" style="font-size: 30px;"></i>
                    </div>
                    <span>{{ user.name }}</span>
                    <span v-if="user.userId === userId" class="badge">TU</span>
                  </div>
                  <input
                    type="checkbox"
                    :value="user.userId"
                    :checked="isSelected(user)"
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
              @click="addMembers"
              :disabled="selectedUsers.length === 0"
            >
              Aggiungi Membri
            </button>
          </div>
        </div>
      </div>
    </div>
  </template>
  
  <script>
  import { searchUsers, getGroupMembers, addToGroup } from "@/services/axios";
  
  export default {
    data() {
      return {
        conversationId: null,
        selectedUsers: [],
        searchResults: [],
        userId: null,
      };
    },
    methods: {
      open(conversationId) {
        this.conversationId = conversationId;
        const modal = new bootstrap.Modal(this.$refs.modal);
        modal.show();
        this.reset();
  
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
        this.selectedUsers = [];
        this.searchResults = [];
      },
      async search() {
        try {
          const allUsers = await searchUsers(this.userId, "");
          const groupMembers = await getGroupMembers(this.userId, this.conversationId);

          // Filtra gli utenti per escludere quelli già nel gruppo
          this.searchResults = allUsers.filter(user => !groupMembers.includes(user.name));
        } catch (error) {
          console.error("Errore durante la ricerca degli utenti:", error);
        }
      },
      isSelected(user) {
        return this.selectedUsers.some(
          (selectedUser) => selectedUser.userId === user.userId
        );
      },
      toggleUserSelection(user) {
        if (user.userId === this.userId) {
          alert("Non puoi aggiungere te stesso. Sei già nel gruppo.");
          return;
        }
        const index = this.selectedUsers.findIndex(
          (selectedUser) => selectedUser.userId === user.userId
        );
        if (index === -1) {
          this.selectedUsers.push(user);
        } else {
          this.selectedUsers.splice(index, 1);
        }
      },
      async addMembers() {
        try {
          const userId = localStorage.getItem("userId");
          for (const user of this.selectedUsers) {
            await addToGroup(userId, this.conversationId, user.name);
          }
          this.close();
        } catch (error) {
          console.error("Errore durante l'aggiunta dei membri:", error);
        }
      },
    },
  };
  </script>
  
<style scoped>
@import '../assets/style.css';
  /* Applica lo stile card alle box principali */
.modal-content, .user-list, .modal-body, .modal-header, .modal-footer {
  background: var(--color-bg-secondary) !important;
  color: var(--color-fg) !important;
  border-radius: 8px;
  border: 1px solid var(--color-border);
}
.card {
  background: var(--color-bg-secondary);
  border: 1px solid var(--color-border);
  border-radius: 8px;
}
.user-photo img {
    width: 70px;
    height: 70px;
    object-fit: cover;
    border-radius: 50%;
  }
</style>
