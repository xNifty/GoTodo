document.addEventListener("DOMContentLoaded", () => {
    setInterval(() => {
        fetch('/api/')
            .then(response => response.json())
            .then(data => {
                const tbody = document.querySelector("tbody");
                tbody.innerHTML = data.map(task => `
                    <tr>
                        <td>${task.ID}</td>
                        <td>${task.Title}</td>
                        <td>${task.Description}</td>
                        <td>${task.Completed ? '<font color="green">Complete</font>' : '<font color="red">Incomplete</font>'}</td>
                    </tr>
                `).join('');
            });
    }, 10000);

    function restartTimer() {
        let time = 10;
        const timer = document.getElementById('time');
        const interval = setInterval(() => {
            time--;
            timer.innerText = time;
            if (time === 0) {
                clearInterval(interval);
                timer.innerText = '10';
                restartTimer();
            }
        }, 1000);
    }
    restartTimer();
});
