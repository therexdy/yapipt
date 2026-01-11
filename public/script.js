document.body.hidden = true;

(function() {
    const getCookie = (name) => {
        const row = document.cookie.split("; ").find(r => r.startsWith(`${name}=`));
        return row ? row.split("=")[1] : null;
    };

    if (getCookie("session_token")) {
        window.location.replace("/chat/index.html");
    } else {
        document.body.hidden = false;
    }
})();

document.addEventListener("DOMContentLoaded", () => {
    const loginForm = document.getElementById("login-form");
    if (!loginForm) return;

    loginForm.addEventListener("submit", async (e) => {
        e.preventDefault();

        const username = document.getElementById("username")?.value;
        const password = document.getElementById("password")?.value;

        const payload = {
            user_name: username,
            password: password
        };

        try {
            const response = await fetch("/api/user", {
                method: "POST",
                headers: { "Content-Type": "application/json" },
                body: JSON.stringify(payload)
            });

            if (response.ok) {
                const result = await response.json();
                if (result.success && result.user_name) {
                    document.cookie = `session_token=${encodeURIComponent(result.session_token)}; path=/; SameSite=Strict`;
                    document.cookie = `user_name=${encodeURIComponent(result.user_name)}; path=/; SameSite=Strict`;
                    window.location.replace("/chat/index.html");
                } else {
                    alert("Unauthorized: Invalid credentials or missing user data");
                }
            }
        } catch (err) {
            console.error(err);
        }
    });
});
