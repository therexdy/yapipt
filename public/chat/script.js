document.body.hidden = true;

document.addEventListener("DOMContentLoaded", () => {
    const messagesEl = document.getElementById("messages");
    const form = document.getElementById("chat-form");
    const input = document.getElementById("message-input");
    const typingEl = document.getElementById("typing-indicator");

    const getCookie = (name) => {
        const row = document.cookie.split("; ").find(r => r.startsWith(`${name}=`));
        return row ? decodeURIComponent(row.split("=")[1]) : null;
    };

    const userName = getCookie("user_name");
    const token = getCookie("session_token");

    if (!userName || !token) {
        document.cookie = "session_token=; Max-Age=0; path=/;";
        document.cookie = "user_name=; Max-Age=0; path=/;";
        window.location.replace("/");
        return;
    }

    let socket = null;
    let typingTimeout = null;
    const typingUsers = new Set();

    const renderMessage = (user, text, type) => {
        const div = document.createElement("div");
        div.className = `message ${type}`;
        const u = document.createElement("span");
        u.className = "username";
        u.textContent = `${user}: `;
        const t = document.createElement("span");
        t.className = "text";
        t.textContent = text;
        div.append(u, t);
        messagesEl.appendChild(div);
        div.scrollIntoView({ behavior: "smooth" });
    };

    const renderSystemMessage = (text) => {
        const div = document.createElement("div");
        div.className = "message system";
        div.textContent = text;
        messagesEl.appendChild(div);
        div.scrollIntoView({ behavior: "smooth" });
    };

    const updateTypingUI = () => {
        const names = Array.from(typingUsers);
        if (names.length === 0) {
            typingEl.classList.add("hidden");
            return;
        }
        typingEl.textContent = names.length === 1 
            ? `${names[0]} is typing...` 
            : `${names[0]} and others typing...`;
        typingEl.classList.remove("hidden");
    };

    const initSocket = () => {
        const protocol = location.protocol === "https:" ? "wss:" : "ws:";
        socket = new WebSocket(`${protocol}//${location.host}/ws?user=${encodeURIComponent(userName)}`);

        socket.onopen = () => {
            document.body.hidden = false;
            document.body.style.opacity = "1";
        };

        socket.onmessage = (e) => {
            let data;
            try { data = JSON.parse(e.data); } catch { return; }

            if (data.type === "msg_data") {
                renderMessage(data.user, data.msg, "incoming");
            } else if (data.type === "msg_indct") {
                if (data.user === userName) return;

                switch (data.indct_type) {
                    case "typing":
                        typingUsers.add(data.user);
                        break;
                    case "stopped_typing":
                    case "left":
                        typingUsers.delete(data.user);
                        break;
                    case "joined":
                        renderSystemMessage(`${data.user} joined the chat`);
                        break;
                }
                updateTypingUI();
            }
        };

        socket.onclose = () => {
            document.cookie = "session_token=; Max-Age=0; path=/;";
            document.cookie = "user_name=; Max-Age=0; path=/;";
            window.location.replace("/");
        };

        socket.onerror = () => socket.close();
    };

    form.addEventListener("submit", e => {
        e.preventDefault();
        const text = input.value.trim();
        if (!text || !socket || socket.readyState !== WebSocket.OPEN) return;
        
        //renderMessage(userName, text, "outgoing");
        socket.send(JSON.stringify({ 
            type: "msg_data", 
            user: userName, 
            msg: text,
            sent_time: new Date().toISOString()
        }));
        input.value = "";
        
        socket.send(JSON.stringify({ type: "msg_indct", indct_type: "stopped_typing", user: userName }));
    });

    input.addEventListener("input", () => {
        if (socket?.readyState === WebSocket.OPEN) {
            socket.send(JSON.stringify({ type: "msg_indct", indct_type: "typing", user: userName }));
        }
        
        clearTimeout(typingTimeout);
        typingTimeout = setTimeout(() => {
            if (socket?.readyState === WebSocket.OPEN) {
                socket.send(JSON.stringify({ type: "msg_indct", indct_type: "stopped_typing", user: userName }));
            }
        }, 1000);
    });

    initSocket();
});
