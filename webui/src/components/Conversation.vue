<template>
	<!-- Componente per una singola conversazione nella lista. -->
	<!-- Sintassi base, commenti in italiano. -->
	<li
	  class="list-group-item list-group-item-action"
	  :class="{ active: isActive }"
	  @click="selectConversation"
	>
	  <div class="d-flex align-items-center">
		<!-- Foto profilo o icona gruppo -->
		<img
		  v-if="conversation.photo"
		  :src="`data:image/png;base64,${conversation.photo}`"
		  alt="Foto profilo"
		  class="rounded-circle me-3"
		  style="width: 40px; height: 40px; object-fit: cover;"
		/>
		<i v-else class="bi bi-person-circle me-3" style="font-size: 30px;"></i>
		
		<div>
		  <!-- Nome conversazione -->
		  <h5 class="mb-1">{{ conversation.name }}</h5>
		  <!-- Anteprima messaggio -->
		  <p class="mb-0 text-truncate">
			{{ getLastMessagePreview(conversation) }}
		  </p>
		  <!-- Timestamp ultimo messaggio -->
		  <small 
		 	:class="{ 'text-muted': !isActive, 'text-white': isActive }" 
		  >{{ formatTimestamp(conversation.lastMessage?.timestamp) }}</small>
		</div>
	  </div>
	</li>
  </template>
  
  <script>
  export default {
	props: {
	  conversation: {
		type: Object,
		required: true,
	  },
	  isActive: {
		type: Boolean,
		default: false,
	  },
	},
	methods: {
	  selectConversation() {
		this.$emit("select", this.conversation);
	  },
	  getLastMessagePreview(conversation) {
		if (conversation.lastMessage) {
			const lastMessage = conversation.lastMessage;
			// Se l'ultimo messaggio è una foto, restituisci un'icona o un messaggio specifico
			if (!lastMessage.preview) {
			return "📸 Messaggio foto";
			}
			// Trunca l'anteprima se è più lunga di 28 caratteri
			if (lastMessage.preview.length > 28) {
				return lastMessage.preview.substring(0, 28) + '...';
			}
			return lastMessage.preview;
		}
		return 'Nessun messaggio ancora';
	  },
	  formatTimestamp(timestamp) {
		if (timestamp) {
			const date = new Date(timestamp);
			return date.toLocaleString(); // Formatta il timestamp in una data leggibile
		}
		return 'Nessun timestamp';
	  },
	},
  };
  </script>
  
  <style scoped>
  @import '../assets/style.css';
  .list-group-item {
	cursor: pointer;
  }
  .list-group-item.active {
	background-color: #007bff;
	color: #fff;
  }
  .text-truncate {
	overflow: hidden;
	white-space: nowrap;
	text-overflow: ellipsis;
  }
  </style>
