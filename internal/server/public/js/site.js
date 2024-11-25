//document.addEventListener("DOMContentLoaded", () => {
//    const timer = document.getElementById('time');
//    let time = 10;
//
//    const tbody = document.querySelector("tbody");
//
//    function updateTable() {
//        fetch('/api/fetch-tasks')
//            .then(response => response.text())
//            .then(html => {
//                tbody.innerHTML = html;
//            })
//            .catch(error => {
//                console.error(error);
//                tbody.innerHTML = `
//                    <tr>
//                        <td colspan="4">No tasks available</td>
//                    </tr>
//                `;
//            });
//    }
//
//    function restartTimer() {
//        time = 10;
//        const interval = setInterval(() => {
//            time--;
//            timer.innerText = time;
//            if (time === 0) {
//                clearInterval(interval);
//                updateTable();
//                restartTimer();
//            }
//        }, 1000);
//    }
//    updateTable();
//    restartTimer();
//});
