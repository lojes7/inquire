import { useState } from "react";
import Sidebar from "../components/Sidebar";
import "../styles/CalendarPage.css";

export default function CalendarPage() {
  const today = new Date();

  const todayDate = today.getDate();
  const todayMonth = today.getMonth();
  const todayYear = today.getFullYear();

  const [currentDate, setCurrentDate] = useState(today);
  const [selectedDay, setSelectedDay] = useState(todayDate);

  const [tasks, setTasks] = useState({
    10: [{ text: "产品调研讨论", deadline: "" }],
    15: [{ text: "数据库优化", deadline: "" }],
    20: [
      { text: "项目答辩", deadline: "" },
      { text: "材料准备", deadline: "" },
    ],
  });

  const [newTask, setNewTask] = useState("");
  const [newDate, setNewDate] = useState(todayDate);
  const [deadline, setDeadline] = useState("");

  // 弹窗控制
  const [showAddModal, setShowAddModal] = useState(false);
  const [showEditModal, setShowEditModal] = useState(false);
  const [showMoreModal, setShowMoreModal] = useState(false);

  const [editIndex, setEditIndex] = useState(null);
  const [editText, setEditText] = useState("");
  const [editDeadline, setEditDeadline] = useState("");

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
  添加任务（弹窗）
  =========================
  */
  const handleAddTask = () => {
    if (!newTask.trim()) return;

    setTasks((prev) => ({
      ...prev,
      [newDate]: [
        ...(prev[newDate] || []),
        { text: newTask, deadline },
      ],
    }));

    setNewTask("");
    setDeadline("");
    setShowAddModal(false);
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
  打开编辑弹窗
  =========================
  */
  const openEdit = (index) => {
    const task = tasks[selectedDay][index];
    setEditIndex(index);
    setEditText(task.text);
    setEditDeadline(task.deadline);
    setShowEditModal(true);
  };

  /*
  =========================
  保存编辑
  =========================
  */
  const saveEdit = () => {
    setTasks((prev) => {
      const updated = [...prev[selectedDay]];
      updated[editIndex] = {
        text: editText,
        deadline: editDeadline,
      };

      return {
        ...prev,
        [selectedDay]: updated,
      };
    });

    setShowEditModal(false);
  };

  /*
  =========================
  总任务统计
  =========================
  */
  const totalTasks = Object.values(tasks).flat().length;

  return (
    <div className="calendar-layout">
      <Sidebar />

      <div className="calendar-page">
        {/* 日历 */}
        <div className="calendar-main">
          <div className="calendar-top">
            <h2>📅 智能日历计划</h2>

            <div className="calendar-switch">
              <button onClick={() => changeMonth(-1)}>←</button>
              <span>
                {currentDate.getFullYear()} 年{" "}
                {currentDate.getMonth() + 1} 月
              </span>
              <button onClick={() => changeMonth(1)}>→</button>
            </div>
          </div>

          <div className="calendar-grid">
            {Array.from({ length: daysInMonth }).map((_, i) => {
              const day = i + 1;
              const dayTasks = tasks[day] || [];

              const isToday =
                day === todayDate &&
                currentDate.getMonth() === todayMonth &&
                currentDate.getFullYear() === todayYear;

              return (
                <div
                  key={i}
                  className={`day-box 
                    ${selectedDay === day ? "selected" : ""}
                    ${isToday ? "today" : ""}
                  `}
                  onClick={() => setSelectedDay(day)}
                >
                  <div className="day-number">{day}</div>

                  {dayTasks.length > 0 && (
                    <div className="mini-task">
                      <p>{dayTasks[0].text}</p>

                      {dayTasks.length > 1 && (
                        <span
                          onClick={(e) => {
                            e.stopPropagation();
                            setShowMoreModal(true);
                          }}
                        >
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

        {/* 右侧 */}
        <div className="right-wrapper">
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

            <button
              className="add-btn"
              onClick={() => setShowAddModal(true)}
            >
              <span className="plus">＋</span>
              <span>添加任务</span>
            </button>
          </div>

          {/* 任务详情 */}
          <div className="detail-panel">
            <h2>📌 {selectedDay} 日任务</h2>

            <div className="task-list">
              {tasks[selectedDay]?.length ? (
                tasks[selectedDay].map((task, index) => (
                  <div key={index} className="task-item">
                    <div>
                      <span>{task.text}</span>
                      {task.deadline && (
                        <p className="deadline">
                          ⏰ 截止：{task.deadline}
                        </p>
                      )}
                    </div>

                    <div>
                      <button
                        onClick={() => openEdit(index)}
                      >
                        编辑
                      </button>

                      <button
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
                <p>暂无任务</p>
              )}
            </div>
          </div>
        </div>
      </div>

      {/* =========================
          添加任务弹窗
      ========================= */}
      {showAddModal && (
        <div className="modal">
          <div className="modal-content fancy">
            <h3>✨ 新建任务</h3>

            <div className="form-group">
              <label>任务内容</label>
              <input
                placeholder="请输入任务..."
                value={newTask}
                onChange={(e) => setNewTask(e.target.value)}
              />
            </div>

            <div className="form-group">
              <label>选择日期</label>
              <input
                type="number"
                min="1"
                max={daysInMonth}
                value={newDate}
                onChange={(e) => setNewDate(e.target.value)}
              />
            </div>

            <div className="form-group">
              <label>截止时间</label>
              <input
                type="datetime-local"
                value={deadline}
                onChange={(e) => setDeadline(e.target.value)}
              />
            </div>

            <div className="modal-actions">
              <button className="confirm" onClick={handleAddTask}>
                ✔ 添加
              </button>

              <button
                className="cancel"
                onClick={() => setShowAddModal(false)}
              >
                取消
              </button>
            </div>
          </div>
        </div>
      )}

      {/* =========================
          编辑弹窗
      ========================= */}
      {showEditModal && (
        <div className="modal">
          <div className="modal-content fancy">
            <h3>✏️ 编辑任务</h3>

            <div className="form-group">
              <label>任务内容</label>
              <input
                value={editText}
                onChange={(e) =>
                  setEditText(e.target.value)
                }
              />
            </div>

            <div className="form-group">
              <label>截止时间</label>
              <input
                type="datetime-local"
                value={editDeadline}
                onChange={(e) =>
                  setEditDeadline(e.target.value)
                }
              />
            </div>

            <div className="modal-actions">
              <button className="confirm" onClick={saveEdit}>
                ✔ 保存修改
              </button>

              <button
                className="cancel"
                onClick={() => setShowEditModal(false)}
              >
                取消
              </button>
            </div>
          </div>
        </div>
      )}

      {/* =========================
          查看更多弹窗
      ========================= */}
     {showMoreModal && (
        <div className="modal">
          <div className="modal-content fancy large">
            <h3>📋 {selectedDay} 日全部任务</h3>

            <div className="task-modal-list">
              {tasks[selectedDay]?.map((t, i) => (
                <div key={i} className="task-card">
                  <div className="task-info">
                    <span className="task-text">
                      {t.text}
                    </span>

                    {t.deadline && (
                      <span className="task-deadline">
                        ⏰ {t.deadline}
                      </span>
                    )}
                  </div>

                  <div className="task-actions">
                    <button
                      className="mini edit"
                      onClick={() => {
                        setShowMoreModal(false);
                        openEdit(i);
                      }}
                    >
                      编辑
                    </button>

                    <button
                      className="mini delete"
                      onClick={() => deleteTask(i)}
                    >
                      删除
                    </button>
                  </div>
                </div>
              ))}
            </div>

            <button
              className="close-btn"
              onClick={() => setShowMoreModal(false)}
            >
              关闭
            </button>
          </div>
        </div>
      )}
    </div>
  );
}