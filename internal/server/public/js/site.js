document.addEventListener("DOMContentLoaded", () => {
    const timer = document.getElementById('time');
    let time = 10;

    const tbody = document.querySelector("tbody");

    function updateTable() {
        fetch('/api/')
            .then(response => response.json())
            .then(data => {
                tbody.innerHTML = data.map(task => `
                    <tr>
                        <td>${task.ID}</td>
                        <td>${task.Title}</td>
                        <td>${task.Description}</td>
                        <td>${task.Completed ? '<font color="green">Complete</font>' : '<font color="red">Incomplete</font>'}</td>
                    </tr>
                `).join('');
            })
            .catch(error => {
                tbody.innerHTML = `
                    <tr>
                        <td colspan="4">No tasks available</td>
                    </tr>
                `;
            });
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
