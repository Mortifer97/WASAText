<template>
  <!-- Dialog per cambiare il nome del gruppo. -->
  <!-- Sintassi base, commenti in italiano. -->
  <div v-if="isOpen" class="modal fade show d-block" tabindex="-1" role="dialog" aria-labelledby="changeGroupNameLabel" aria-hidden="true">
    <div class="modal-dialog" role="document">
      <div class="modal-content">
        <div class="modal-header">
          <h5 class="modal-title" id="changeGroupNameLabel">Cambia Nome Gruppo</h5>
          <button type="button" class="btn-close" aria-label="Close" @click="closeDialog"></button>
        </div>
        <div class="modal-body">
          <form @submit.prevent="submitGroupNameChange">
            <div class="mb-3">
              <label for="groupNameInput" class="form-label">Nuovo Nome Gruppo</label>
              <input
                type="text"
                id="groupNameInput"
                class="form-control"
                v-model="newGroupName"
                placeholder="Inserisci il nuovo nome del gruppo"
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
  props: {
    conversation: Object,
  },
  emits: ['group-name-changed'],
  data() {
    return {
      isOpen: false,
      newGroupName: '',
    };
  },
  methods: {
    open() {
      this.newGroupName = this.conversation?.name || '';
      this.isOpen = true;
    },
    closeDialog() {
      this.isOpen = false;
      this.newGroupName = '';
    },
    submitGroupNameChange() {
      if (this.newGroupName.trim() === "") {
        alert("Il nome del gruppo non può essere vuoto.");
        return;
      }
      this.$emit('group-name-changed', this.newGroupName.trim());
      this.closeDialog();
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

.modal-backdrop {
  z-index: 1040;
}

.modal {
  z-index: 1050;
}
</style>
