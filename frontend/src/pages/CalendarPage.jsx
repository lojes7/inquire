import { useState } from "react";
import Sidebar from "../components/Sidebar";
import "../styles/CalendarPage.css";

export default function CalendarPage() {
  const today = new Date();

  const [currentDate, setCurrentDate] = useState(today);
  const [selectedDay, setSelectedDay] = useState(today.getDate());

  const [tasks, setTasks] = useState({
    10: ["产品调研讨论", "UI设计会议", "写日报"],
    15: ["数据库优化"],
    20: ["项目答辩", "材料准备"],
  });

  const [newTask, setNewTask] = useState("");

  /*
  =========================
  获取当月天数
  =========================
  */
  const getDaysInMonth = (date) => {
    return new Date(
      date.getFullYear(),
      date.getMonth() + 1,
      0
    ).getDate();
  };

  const daysInMonth = getDaysInMonth(currentDate);

  /*
  =========================
  切换月份
  =========================
  */
  const changeMonth = (offset) => {
    setCurrentDate(
      new Date(
        currentDate.getFullYear(),
        currentDate.getMonth() + offset,
        1
      )
    );

    setSelectedDay(1);
  };

  /*
  =========================
  添加任务
  =========================
  */
  const addTask = () => {
    if (!newTask.trim()) return;

    setTasks((prev) => ({
      ...prev,
      [selectedDay]: [...(prev[selectedDay] || []), newTask],
    }));

    setNewTask("");
  };

  /*
  =========================
  删除任务
  =========================
  */
  const deleteTask = (index) => {
    setTasks((prev) => {
      const updated = [...(prev[selectedDay] || [])];

      updated.splice(index, 1);

      return {
        ...prev,
        [selectedDay]: updated,
      };
    });
  };

  /*
  =========================
  编辑任务
  =========================
  */
  const editTask = (index) => {
    const newText = prompt(
      "编辑任务",
      tasks[selectedDay][index]
    );

    if (!newText) return;

    setTasks((prev) => {
      const updated = [...prev[selectedDay]];

      updated[index] = newText;

      return {
        ...prev,
        [selectedDay]: updated,
      };
    });
  };

  /*
  =========================
  总任务统计
  =========================
  */
  const totalTasks = Object.values(tasks).flat().length;

  return (
    <div className="calendar-layout">

      {/* 左侧 Sidebar */}
      <Sidebar />

      {/* 页面主体 */}
      <div className="calendar-page">

        {/* 日历区域 */}
        <div className="calendar-main">

          <div className="calendar-top">
            <h2>📅 智能日历计划</h2>

            <div className="calendar-switch">
              <button onClick={() => changeMonth(-1)}>←</button>

              <span>
                {currentDate.getFullYear()} 年
                {currentDate.getMonth() + 1} 月
              </span>

              <button onClick={() => changeMonth(1)}>→</button>
            </div>
          </div>

          <div className="calendar-grid">
            {Array.from({ length: daysInMonth }).map((_, i) => {
              const day = i + 1;

              const dayTasks = tasks[day] || [];

              return (
                <div
                  key={i}
                  className={`day-box ${
                    selectedDay === day ? "selected" : ""
                  }`}
                  onClick={() => setSelectedDay(day)}
                >
                  <div className="day-number">{day}</div>

                  {dayTasks.length > 0 && (
                    <div className="mini-task">
                      <p>{dayTasks[0]}</p>

                      {dayTasks.length > 1 && (
                        <span>
                          +{dayTasks.length - 1} 更多
                        </span>
                      )}
                    </div>
                  )}
                </div>
              );
            })}
          </div>

        </div>

        {/* 右侧功能区 */}
        <div className="right-wrapper">

          {/* 智能管理 */}
          <div className="smart-panel">

            <h2>🤖 智能日程管理</h2>

            <div className="smart-card">
              <p>总任务数</p>
              <strong>{totalTasks}</strong>
            </div>

            <div className="smart-card">
              <p>今日任务</p>
              <strong>
                {tasks[selectedDay]?.length || 0}
              </strong>
            </div>

            <div className="smart-card">
              <p>完成率建议</p>
              <strong>
                {Math.min(
                  100,
                  Math.round(
                    (selectedDay / daysInMonth) * 100
                  )
                )}
                %
              </strong>
            </div>

            <div className="ai-tip">
              💡 AI建议：
              <br />
              今天适合优先完成高优先级任务，
              建议控制在 3 项以内提高效率。
            </div>

          </div>

          {/* 任务详情 */}
          <div className="detail-panel">

            <h2>📌 {selectedDay} 日任务详情</h2>

            <div className="task-list">
              {tasks[selectedDay]?.length ? (
                tasks[selectedDay].map((task, index) => (
                  <div
                    key={index}
                    className="task-item"
                  >
                    <span>{task}</span>

                    <div>
                      <button
                        className="edit-btn"
                        onClick={() =>
                          editTask(index)
                        }
                      >
                        编辑
                      </button>

                      <button
                        className="delete-btn"
                        onClick={() =>
                          deleteTask(index)
                        }
                      >
                        删除
                      </button>
                    </div>
                  </div>
                ))
              ) : (
                <p className="empty-tip">
                  暂无任务
                </p>
              )}
            </div>

            <div className="add-task">
              <input
                value={newTask}
                onChange={(e) =>
                  setNewTask(e.target.value)
                }
                placeholder="新增任务..."
              />

              <button onClick={addTask}>
                添加
              </button>
            </div>

          </div>

        </div>

      </div>
    </div>
  );
}