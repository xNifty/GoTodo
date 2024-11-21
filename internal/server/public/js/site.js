document.addEventListener("DOMContentLoaded", () => {
    const timer = document.getElementById('time');
    let time = 10;

    function updateTable() {
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
            })
            .catch(error => console.error(error));
    }

    function restartTimer() {
        time = 10;
        const interval = setInterval(() => {
            time--;
            timer.innerText = time;
            if (time === 0) {
                clearInterval(interval);
                updateTable();
                restartTimer();
            }
        }, 1000);
    }
    updateTable();
    restartTimer();
});
