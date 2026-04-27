import { describe, expect, it } from 'vitest';
import { defaultTaskId, findTaskById, internshipTasks } from './tasks';

describe('internshipTasks', () => {
  it('contains exactly Day 0 through Day 9', () => {
    expect(internshipTasks.map((task) => task.day)).toEqual([0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13]);
  });

  it('uses stable unique ids and complete task content', () => {
    const ids = internshipTasks.map((task) => task.id);

    expect(new Set(ids).size).toBe(ids.length);

    for (const task of internshipTasks) {
      expect(task.title.length).toBeGreaterThanOrEqual(4);
      expect(task.summary.length).toBeGreaterThan(8);
      expect(task.background.length).toBeGreaterThan(20);
      expect(task.objective.length).toBeGreaterThan(10);
      expect(task.starterCode).toContain('package main');
      expect(task.criteria.length).toBeGreaterThanOrEqual(3);
      expect(task.lesson.length).toBeGreaterThanOrEqual(3);
      expect(task.mentorHints.length).toBeGreaterThanOrEqual(3);
      expect(task.review.length).toBeGreaterThanOrEqual(3);
      expect(task.checks.length).toBeGreaterThanOrEqual(2);
    }
  });

  it('finds tasks by id and falls back to the default task', () => {
    expect(defaultTaskId).toBe('day-0-first-run');
    expect(findTaskById('day-3-validation').day).toBe(3);
    expect(findTaskById('missing-task').id).toBe(defaultTaskId);
  });
});
