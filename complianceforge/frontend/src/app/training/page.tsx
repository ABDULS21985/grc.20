'use client';

import { useEffect, useState, useCallback } from 'react';
import api from '@/lib/api';

const STATUS_COLORS: Record<string, string> = {
  assigned: 'bg-blue-100 text-blue-800',
  in_progress: 'bg-yellow-100 text-yellow-800',
  completed: 'bg-green-100 text-green-800',
  failed: 'bg-red-100 text-red-800',
  overdue: 'bg-red-100 text-red-800',
  exempted: 'bg-gray-100 text-gray-600',
};

const CATEGORY_LABELS: Record<string, string> = {
  security_awareness: 'Security Awareness',
  privacy: 'Privacy',
  compliance: 'Compliance',
  technical: 'Technical',
  management: 'Management',
  onboarding: 'Onboarding',
  custom: 'Custom',
};

interface TrainingAssignment {
  id: string;
  programme_id: string;
  programme_name: string;
  programme_category: string;
  status: string;
  assigned_at: string;
  due_date: string | null;
  started_at: string | null;
  completed_at: string | null;
  score: number | null;
  passed: boolean | null;
  attempts: number;
  time_spent_minutes: number;
  certificate_url: string;
}

interface QuizQuestion {
  id: string;
  question_text: string;
  answer_options: { text: string }[];
  correct_answer_index: number;
  explanation: string;
}

type Tab = 'assignments' | 'quiz';

export default function TrainingPage() {
  const [tab, setTab] = useState<Tab>('assignments');
  const [assignments, setAssignments] = useState<TrainingAssignment[]>([]);
  const [loading, setLoading] = useState(true);
  const [startingId, setStartingId] = useState<string | null>(null);

  // Quiz state
  const [quizAssignment, setQuizAssignment] = useState<TrainingAssignment | null>(null);
  const [quizQuestions, setQuizQuestions] = useState<QuizQuestion[]>([]);
  const [currentQuestion, setCurrentQuestion] = useState(0);
  const [selectedAnswers, setSelectedAnswers] = useState<Record<number, number>>({});
  const [quizSubmitted, setQuizSubmitted] = useState(false);
  const [quizScore, setQuizScore] = useState(0);

  const loadAssignments = useCallback(async () => {
    setLoading(true);
    try {
      const res = await api.get('/training/my-assignments');
      setAssignments(res.data?.data || []);
    } catch {
      /* ignore */
    }
    setLoading(false);
  }, []);

  useEffect(() => {
    loadAssignments();
  }, [loadAssignments]);

  const handleStart = async (assignment: TrainingAssignment) => {
    setStartingId(assignment.id);
    try {
      await api.post(`/training/assignments/${assignment.id}/start`);
      await loadAssignments();
    } catch {
      /* ignore */
    }
    setStartingId(null);
  };

  const handleStartQuiz = async (assignment: TrainingAssignment) => {
    setQuizAssignment(assignment);
    setTab('quiz');
    setCurrentQuestion(0);
    setSelectedAnswers({});
    setQuizSubmitted(false);
    setQuizScore(0);

    // Load quiz questions for this programme
    try {
      const res = await api.get(`/training/programmes/${assignment.programme_id}`);
      // In a real implementation, content would be fetched separately
      // For now, simulate quiz questions from the programme content
      const programme = res.data?.data;
      if (programme) {
        // Placeholder: generate sample quiz UI
        setQuizQuestions([]);
      }
    } catch {
      /* ignore */
    }
  };

  const handleSelectAnswer = (questionIndex: number, answerIndex: number) => {
    if (quizSubmitted) return;
    setSelectedAnswers((prev) => ({ ...prev, [questionIndex]: answerIndex }));
  };

  const handleSubmitQuiz = async () => {
    if (!quizAssignment) return;

    // Calculate score
    let correct = 0;
    quizQuestions.forEach((q, i) => {
      if (selectedAnswers[i] === q.correct_answer_index) {
        correct++;
      }
    });
    const score = quizQuestions.length > 0 ? Math.round((correct / quizQuestions.length) * 100) : 100;
    setQuizScore(score);
    setQuizSubmitted(true);

    // Submit completion
    try {
      await api.post(`/training/assignments/${quizAssignment.id}/complete`, {
        score,
        time_spent_minutes: 15,
      });
      await loadAssignments();
    } catch {
      /* ignore */
    }
  };

  const handleDownloadCertificate = (assignment: TrainingAssignment) => {
    if (assignment.certificate_url) {
      window.open(assignment.certificate_url, '_blank');
    }
  };

  const pendingAssignments = assignments.filter(
    (a) => a.status === 'assigned' || a.status === 'in_progress' || a.status === 'overdue'
  );
  const completedAssignments = assignments.filter(
    (a) => a.status === 'completed' || a.status === 'failed' || a.status === 'exempted'
  );

  const tabs = [
    { key: 'assignments' as const, label: 'My Assignments' },
    { key: 'quiz' as const, label: 'Training / Quiz', disabled: !quizAssignment },
  ];

  return (
    <div>
      <div className="flex items-center justify-between mb-8">
        <div>
          <h1 className="text-2xl font-bold text-gray-900">Training Hub</h1>
          <p className="text-gray-500 mt-1">
            Complete your assigned training programmes and track your progress
          </p>
        </div>
        <a
          href="/training/admin"
          className="px-4 py-2 bg-indigo-600 text-white text-sm rounded-lg hover:bg-indigo-700 transition-colors"
        >
          Admin Panel
        </a>
      </div>

      {/* Summary Cards */}
      <div className="grid grid-cols-1 md:grid-cols-4 gap-4 mb-6">
        <div className="card p-4">
          <p className="text-xs text-gray-500 uppercase tracking-wide">Total Assigned</p>
          <p className="text-2xl font-bold text-gray-900 mt-1">{assignments.length}</p>
        </div>
        <div className="card p-4">
          <p className="text-xs text-gray-500 uppercase tracking-wide">Pending</p>
          <p className="text-2xl font-bold text-yellow-600 mt-1">{pendingAssignments.length}</p>
        </div>
        <div className="card p-4">
          <p className="text-xs text-gray-500 uppercase tracking-wide">Completed</p>
          <p className="text-2xl font-bold text-green-600 mt-1">
            {assignments.filter((a) => a.status === 'completed').length}
          </p>
        </div>
        <div className="card p-4">
          <p className="text-xs text-gray-500 uppercase tracking-wide">Overdue</p>
          <p className="text-2xl font-bold text-red-600 mt-1">
            {assignments.filter((a) => a.status === 'overdue').length}
          </p>
        </div>
      </div>

      <div className="flex gap-1 border-b border-gray-200 mb-6 overflow-x-auto">
        {tabs.map((t) => (
          <button
            key={t.key}
            onClick={() => !t.disabled && setTab(t.key)}
            disabled={t.disabled}
            className={`px-4 py-2.5 text-sm font-medium border-b-2 whitespace-nowrap transition-colors ${
              tab === t.key
                ? 'border-indigo-600 text-indigo-600'
                : t.disabled
                  ? 'border-transparent text-gray-300 cursor-not-allowed'
                  : 'border-transparent text-gray-500 hover:text-gray-700'
            }`}
          >
            {t.label}
          </button>
        ))}
      </div>

      {loading ? (
        <div className="card animate-pulse h-96" />
      ) : tab === 'assignments' ? (
        <div>
          {/* Pending Assignments */}
          {pendingAssignments.length > 0 && (
            <div className="mb-8">
              <h2 className="text-lg font-semibold text-gray-800 mb-4">Pending Training</h2>
              <div className="space-y-3">
                {pendingAssignments.map((a) => (
                  <div key={a.id} className="card p-4 flex items-center justify-between">
                    <div className="flex-1">
                      <div className="flex items-center gap-3">
                        <h3 className="font-medium text-gray-900">{a.programme_name}</h3>
                        <span
                          className={`text-xs px-2 py-0.5 rounded-full ${STATUS_COLORS[a.status] || 'bg-gray-100 text-gray-600'}`}
                        >
                          {a.status.replace('_', ' ')}
                        </span>
                      </div>
                      <div className="flex items-center gap-4 mt-1 text-sm text-gray-500">
                        <span>{CATEGORY_LABELS[a.programme_category] || a.programme_category}</span>
                        {a.due_date && (
                          <span>
                            Due: {new Date(a.due_date).toLocaleDateString()}
                          </span>
                        )}
                        {a.attempts > 0 && <span>Attempts: {a.attempts}</span>}
                        {a.score !== null && <span>Last Score: {a.score}%</span>}
                      </div>
                    </div>
                    <div className="flex gap-2">
                      {a.status === 'assigned' && (
                        <button
                          onClick={() => handleStart(a)}
                          disabled={startingId === a.id}
                          className="px-4 py-2 bg-indigo-600 text-white text-sm rounded-lg hover:bg-indigo-700 disabled:opacity-50"
                        >
                          {startingId === a.id ? 'Starting...' : 'Start Training'}
                        </button>
                      )}
                      {a.status === 'in_progress' && (
                        <button
                          onClick={() => handleStartQuiz(a)}
                          className="px-4 py-2 bg-green-600 text-white text-sm rounded-lg hover:bg-green-700"
                        >
                          Continue / Take Quiz
                        </button>
                      )}
                      {a.status === 'overdue' && (
                        <button
                          onClick={() => handleStart(a)}
                          disabled={startingId === a.id}
                          className="px-4 py-2 bg-red-600 text-white text-sm rounded-lg hover:bg-red-700 disabled:opacity-50"
                        >
                          Start Now (Overdue)
                        </button>
                      )}
                    </div>
                  </div>
                ))}
              </div>
            </div>
          )}

          {/* Completed Assignments */}
          {completedAssignments.length > 0 && (
            <div>
              <h2 className="text-lg font-semibold text-gray-800 mb-4">Completed Training</h2>
              <div className="card overflow-x-auto">
                <table className="w-full text-sm">
                  <thead>
                    <tr className="table-header">
                      <th className="px-4 py-3 text-left">Programme</th>
                      <th className="px-4 py-3 text-left">Category</th>
                      <th className="px-4 py-3 text-left">Status</th>
                      <th className="px-4 py-3 text-right">Score</th>
                      <th className="px-4 py-3 text-right">Attempts</th>
                      <th className="px-4 py-3 text-left">Completed</th>
                      <th className="px-4 py-3 text-left">Certificate</th>
                    </tr>
                  </thead>
                  <tbody>
                    {completedAssignments.map((a) => (
                      <tr key={a.id} className="border-t border-gray-100 hover:bg-gray-50">
                        <td className="px-4 py-3 font-medium text-gray-900">{a.programme_name}</td>
                        <td className="px-4 py-3 text-gray-500">
                          {CATEGORY_LABELS[a.programme_category] || a.programme_category}
                        </td>
                        <td className="px-4 py-3">
                          <span
                            className={`text-xs px-2 py-0.5 rounded-full ${STATUS_COLORS[a.status] || 'bg-gray-100 text-gray-600'}`}
                          >
                            {a.status.replace('_', ' ')}
                          </span>
                        </td>
                        <td className="px-4 py-3 text-right">
                          {a.score !== null ? `${a.score}%` : '-'}
                        </td>
                        <td className="px-4 py-3 text-right">{a.attempts}</td>
                        <td className="px-4 py-3 text-gray-500">
                          {a.completed_at
                            ? new Date(a.completed_at).toLocaleDateString()
                            : '-'}
                        </td>
                        <td className="px-4 py-3">
                          {a.status === 'completed' && a.certificate_url ? (
                            <button
                              onClick={() => handleDownloadCertificate(a)}
                              className="text-indigo-600 hover:text-indigo-800 text-xs font-medium"
                            >
                              Download
                            </button>
                          ) : (
                            <span className="text-gray-400 text-xs">-</span>
                          )}
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            </div>
          )}

          {assignments.length === 0 && (
            <div className="card p-12 text-center">
              <p className="text-gray-500">No training assignments yet.</p>
              <p className="text-sm text-gray-400 mt-1">
                Your administrator will assign training programmes to you.
              </p>
            </div>
          )}
        </div>
      ) : tab === 'quiz' && quizAssignment ? (
        <div className="max-w-2xl mx-auto">
          <div className="card p-6">
            <h2 className="text-lg font-semibold text-gray-800 mb-2">
              {quizAssignment.programme_name}
            </h2>
            <p className="text-sm text-gray-500 mb-6">
              {CATEGORY_LABELS[quizAssignment.programme_category] || quizAssignment.programme_category}{' '}
              Training
            </p>

            {quizQuestions.length > 0 ? (
              <div>
                {/* Question Progress */}
                <div className="flex items-center gap-2 mb-6">
                  {quizQuestions.map((_, i) => (
                    <div
                      key={i}
                      className={`h-2 flex-1 rounded-full ${
                        i === currentQuestion
                          ? 'bg-indigo-600'
                          : selectedAnswers[i] !== undefined
                            ? 'bg-indigo-300'
                            : 'bg-gray-200'
                      }`}
                    />
                  ))}
                </div>

                {/* Current Question */}
                <div className="mb-6">
                  <p className="text-sm text-gray-500 mb-2">
                    Question {currentQuestion + 1} of {quizQuestions.length}
                  </p>
                  <p className="text-gray-900 font-medium mb-4">
                    {quizQuestions[currentQuestion].question_text}
                  </p>
                  <div className="space-y-2">
                    {quizQuestions[currentQuestion].answer_options.map((opt, i) => (
                      <button
                        key={i}
                        onClick={() => handleSelectAnswer(currentQuestion, i)}
                        className={`w-full text-left p-3 rounded-lg border transition-colors ${
                          selectedAnswers[currentQuestion] === i
                            ? quizSubmitted
                              ? i === quizQuestions[currentQuestion].correct_answer_index
                                ? 'border-green-500 bg-green-50'
                                : 'border-red-500 bg-red-50'
                              : 'border-indigo-500 bg-indigo-50'
                            : quizSubmitted &&
                                i === quizQuestions[currentQuestion].correct_answer_index
                              ? 'border-green-500 bg-green-50'
                              : 'border-gray-200 hover:border-gray-300'
                        }`}
                      >
                        <span className="text-sm">{opt.text}</span>
                      </button>
                    ))}
                  </div>
                  {quizSubmitted && quizQuestions[currentQuestion].explanation && (
                    <div className="mt-4 p-3 bg-blue-50 rounded-lg">
                      <p className="text-sm text-blue-800">
                        {quizQuestions[currentQuestion].explanation}
                      </p>
                    </div>
                  )}
                </div>

                {/* Navigation */}
                <div className="flex justify-between">
                  <button
                    onClick={() => setCurrentQuestion(Math.max(0, currentQuestion - 1))}
                    disabled={currentQuestion === 0}
                    className="px-4 py-2 text-sm text-gray-600 hover:text-gray-800 disabled:text-gray-300"
                  >
                    Previous
                  </button>
                  {currentQuestion < quizQuestions.length - 1 ? (
                    <button
                      onClick={() => setCurrentQuestion(currentQuestion + 1)}
                      className="px-4 py-2 bg-indigo-600 text-white text-sm rounded-lg hover:bg-indigo-700"
                    >
                      Next
                    </button>
                  ) : !quizSubmitted ? (
                    <button
                      onClick={handleSubmitQuiz}
                      disabled={Object.keys(selectedAnswers).length < quizQuestions.length}
                      className="px-4 py-2 bg-green-600 text-white text-sm rounded-lg hover:bg-green-700 disabled:opacity-50"
                    >
                      Submit Quiz
                    </button>
                  ) : (
                    <div className="text-right">
                      <p className="text-lg font-bold">Score: {quizScore}%</p>
                      <p className={`text-sm ${quizScore >= 80 ? 'text-green-600' : 'text-red-600'}`}>
                        {quizScore >= 80 ? 'Passed!' : 'Please try again.'}
                      </p>
                    </div>
                  )}
                </div>
              </div>
            ) : (
              <div className="text-center py-8">
                <p className="text-gray-500 mb-4">
                  This training programme does not yet have quiz questions configured.
                </p>
                <button
                  onClick={handleSubmitQuiz}
                  className="px-6 py-2 bg-green-600 text-white text-sm rounded-lg hover:bg-green-700"
                >
                  Mark as Completed
                </button>
                {quizSubmitted && (
                  <div className="mt-4">
                    <p className="text-green-600 font-medium">
                      Training marked as complete. Score: {quizScore}%
                    </p>
                    <button
                      onClick={() => { setTab('assignments'); setQuizAssignment(null); }}
                      className="mt-2 text-indigo-600 hover:text-indigo-800 text-sm"
                    >
                      Back to Assignments
                    </button>
                  </div>
                )}
              </div>
            )}
          </div>
        </div>
      ) : null}
    </div>
  );
}
