// Vue.js Chat компонент

const ChatComponent = {
    name: 'Chat',
    setup() {
        const { ref, reactive, onMounted, nextTick } = Vue;

        const conversations = ref([]);
        const selectedFriend = ref(null);
        const messages = ref([]);
        const newMessage = ref('');
        const isLoading = ref(false);
        const unreadCount = ref(0);

        onMounted(() => {
            loadConversations();
            loadUnreadCount();
        });

        async function loadConversations() {
            conversations.value = await SocialStore.loadFriends();
        }

        async function loadUnreadCount() {
            unreadCount.value = await SocialStore.getUnreadCount();
        }

        async function selectFriend(friend) {
            selectedFriend.value = friend;
            isLoading.value = true;
            messages.value = await SocialStore.getMessages(friend.id);
            isLoading.value = false;

            // Прокрутка вниз
            await nextTick();
            scrollToBottom();
        }

        async function sendMessage() {
            if (!newMessage.value.trim() || !selectedFriend.value) return;

            const result = await SocialStore.sendMessage(
                selectedFriend.value.id,
                newMessage.value.trim()
            );

            if (result.success) {
                messages.value.push(result.message);
                newMessage.value = '';
                await nextTick();
                scrollToBottom();
            } else {
                alert('❌ ' + result.error);
            }
        }

        function scrollToBottom() {
            const container = document.querySelector('.messages-container');
            if (container) {
                container.scrollTop = container.scrollHeight;
            }
        }

        function formatTime(dateStr) {
            const date = new Date(dateStr);
            return date.toLocaleTimeString('ru-RU', {
                hour: '2-digit',
                minute: '2-digit'
            });
        }

        function getUnreadForFriend(friendId) {
            // В реальной реализации нужно считать с сервера
            return 0;
        }

        return {
            conversations,
            selectedFriend,
            messages,
            newMessage,
            isLoading,
            unreadCount,
            loadConversations,
            selectFriend,
            sendMessage,
            formatTime,
            getUnreadForFriend
        };
    },
    template: `
        <div class="chat-page">
            <div class="chat-layout">
                <!-- Список чатов -->
                <div class="conversations-panel">
                    <div class="conversations-header">
                        <h3>💬 Сообщения</h3>
                        <span v-if="unreadCount > 0" class="unread-badge">{{ unreadCount }}</span>
                    </div>

                    <div class="conversations-list">
                        <div v-for="friend in conversations" :key="friend.id"
                             class="conversation-item"
                             :class="{ active: selectedFriend?.id === friend.id }"
                             @click="selectFriend(friend)">
                            <div class="conversation-avatar">{{ friend.name.charAt(0) }}</div>
                            <div class="conversation-info">
                                <div class="conversation-name">{{ friend.name }}</div>
                                <div class="conversation-preview">
                                    {{ friend.is_online ? '🟢 Онлайн' : '⚫ Офлайн' }}
                                </div>
                            </div>
                            <span v-if="getUnreadForFriend(friend.id) > 0" class="conversation-unread">
                                {{ getUnreadForFriend(friend.id) }}
                            </span>
                        </div>

                        <div v-if="!conversations.length" class="empty">
                            Нет сообщений. Начни с добавления друзей!
                        </div>
                    </div>
                </div>

                <!-- Чат -->
                <div class="chat-panel">
                    <div v-if="!selectedFriend" class="chat-empty">
                        <div class="empty-icon">💬</div>
                        <p>Выберите чат чтобы начать общение</p>
                    </div>

                    <template v-else>
                        <!-- Заголовок чата -->
                        <div class="chat-header">
                            <div class="chat-user-info">
                                <div class="chat-avatar">{{ selectedFriend.name.charAt(0) }}</div>
                                <div>
                                    <div class="chat-name">{{ selectedFriend.name }}</div>
                                    <div class="chat-status">
                                        {{ selectedFriend.is_online ? '🟢 Онлайн' : '⚫ Офлайн' }}
                                    </div>
                                </div>
                            </div>
                            <div class="chat-actions">
                                <button class="btn-profile" title="Профиль">👤</button>
                                <button class="btn-challenge" title="Вызвать">⚔️</button>
                            </div>
                        </div>

                        <!-- Сообщения -->
                        <div v-if="isLoading" class="loading-messages">
                            <div class="spinner"></div>
                        </div>

                        <div v-else class="messages-container">
                            <div v-for="msg in messages" :key="msg.id"
                                 class="message"
                                 :class="{ 'message-sent': msg.sender_id !== selectedFriend.id }">
                                <div class="message-bubble">
                                    <div class="message-content">{{ msg.content }}</div>
                                    <div class="message-time">{{ formatTime(msg.created_at) }}</div>
                                </div>
                            </div>

                            <div v-if="!messages.length" class="empty-messages">
                                <p>Нет сообщений. Напиши первым!</p>
                            </div>
                        </div>

                        <!-- Ввод сообщения -->
                        <div class="message-input-container">
                            <input
                                v-model="newMessage"
                                type="text"
                                placeholder="Напишите сообщение..."
                                class="message-input"
                                @keyup.enter="sendMessage"
                            />
                            <button class="send-btn" @click="sendMessage">
                                📤
                            </button>
                        </div>
                    </template>
                </div>
            </div>
        </div>
    `
};

window.ChatComponent = ChatComponent;
