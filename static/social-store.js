// Vue.js Social Store для социальных функций

const SocialStore = {
    state: {
        friends: [],
        friendRequests: [],
        messages: {},
        unreadCount: 0,
        challenges: [],
        activity: [],
        isLoading: false
    },

    // Инициализация
    init() {
        console.log('👥 SocialStore initialized');
    },

    // === Friends ===

    // Отправить запрос в друзья
    async sendFriendRequest(friendId, friendName) {
        try {
            const response = await fetch('/api/social/friends/requests/send', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ friend_id: friendId })
            });

            if (!response.ok) {
                const data = await response.json();
                throw new Error(data.error);
            }

            return { success: true };
        } catch (error) {
            return { success: false, error: error.message };
        }
    },

    // Получить входящие запросы
    async getFriendRequests() {
        try {
            const response = await fetch('/api/social/friends/requests');
            const data = await response.json();
            this.state.friendRequests = data.requests;
            return data.requests;
        } catch (error) {
            console.error('Failed to get friend requests:', error);
            return [];
        }
    },

    // Принять запрос
    async acceptFriendRequest(requestId) {
        try {
            const response = await fetch('/api/social/friends/requests/accept', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ request_id: requestId })
            });

            if (!response.ok) {
                const data = await response.json();
                throw new Error(data.error);
            }

            await this.loadFriends();
            await this.getFriendRequests();
            return { success: true };
        } catch (error) {
            return { success: false, error: error.message };
        }
    },

    // Отклонить запрос
    async rejectFriendRequest(requestId) {
        try {
            const response = await fetch('/api/social/friends/requests/reject', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ request_id: requestId })
            });

            if (!response.ok) {
                const data = await response.json();
                throw new Error(data.error);
            }

            await this.getFriendRequests();
            return { success: true };
        } catch (error) {
            return { success: false, error: error.message };
        }
    },

    // Загрузить друзей
    async loadFriends() {
        this.state.isLoading = true;
        try {
            const response = await fetch('/api/social/friends');
            const data = await response.json();
            this.state.friends = data.friends;
            return data.friends;
        } catch (error) {
            console.error('Failed to load friends:', error);
            return [];
        } finally {
            this.state.isLoading = false;
        }
    },

    // Удалить друга
    async removeFriend(friendId) {
        try {
            const response = await fetch('/api/social/friends/remove', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ friend_id: friendId })
            });

            if (!response.ok) {
                const data = await response.json();
                throw new Error(data.error);
            }

            await this.loadFriends();
            return { success: true };
        } catch (error) {
            return { success: false, error: error.message };
        }
    },

    // === Messages ===

    // Отправить сообщение
    async sendMessage(receiverId, content) {
        try {
            const response = await fetch('/api/social/messages/send', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ receiver_id: receiverId, content })
            });

            if (!response.ok) {
                const data = await response.json();
                throw new Error(data.error);
            }

            const message = await response.json();

            // Добавляем в локальный кэш
            if (!this.state.messages[receiverId]) {
                this.state.messages[receiverId] = [];
            }
            this.state.messages[receiverId].push(message);

            return { success: true, message };
        } catch (error) {
            return { success: false, error: error.message };
        }
    },

    // Получить сообщения
    async getMessages(friendId) {
        try {
            const response = await fetch(`/api/social/messages?friend_id=${friendId}`);
            const data = await response.json();
            this.state.messages[friendId] = data.messages;
            return data.messages;
        } catch (error) {
            console.error('Failed to get messages:', error);
            return [];
        }
    },

    // Получить количество непрочитанных
    async getUnreadCount() {
        try {
            const response = await fetch('/api/social/messages/unread');
            const data = await response.json();
            this.state.unreadCount = data.count;
            return data.count;
        } catch (error) {
            return 0;
        }
    },

    // === Challenges ===

    // Отправить вызов
    async sendChallenge(receiverId, receiverName) {
        try {
            const response = await fetch('/api/social/challenges/send', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ receiver_id: receiverId, receiver_name: receiverName })
            });

            if (!response.ok) {
                const data = await response.json();
                throw new Error(data.error);
            }

            return { success: true };
        } catch (error) {
            return { success: false, error: error.message };
        }
    },

    // Получить вызовы
    async getChallenges(status = 'pending') {
        try {
            const response = await fetch(`/api/social/challenges?status=${status}`);
            const data = await response.json();
            this.state.challenges = data.challenges;
            return data.challenges;
        } catch (error) {
            console.error('Failed to get challenges:', error);
            return [];
        }
    },

    // === Activity ===

    // Получить ленту активности
    async getActivityFeed() {
        try {
            const response = await fetch('/api/social/activity');
            const data = await response.json();
            this.state.activity = data.activities;
            return data.activities;
        } catch (error) {
            console.error('Failed to get activity:', error);
            return [];
        }
    },

    // === Helpers ===

    // Получить друга по ID
    getFriendById(friendId) {
        return this.state.friends.find(f => f.id === friendId);
    },

    // Получить онлайн друзей
    getOnlineFriends() {
        return this.state.friends.filter(f => f.is_online);
    },

    // Получить общее количество друзей
    getFriendsCount() {
        return this.state.friends.length;
    },

    // Получить общее количество запросов
    getRequestsCount() {
        return this.state.friendRequests.length;
    }
};

window.SocialStore = SocialStore;
