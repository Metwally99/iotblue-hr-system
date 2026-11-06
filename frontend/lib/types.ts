import type { Role } from '@/lib/api';

export type Pagination = {
  page: number;
  page_size: number;
  total_items: number;
  total_pages: number;
};

export type ListResponse<T> = { data: T[]; meta: Pagination };

export type Department = {
  id: string;
  name: string;
  created_at: string;
  updated_at: string;
};

export type Employee = {
  id: string;
  department_id: string;
  manager_id: string | null;
  name: string;
  email: string;
  role: Role;
  is_active: boolean;
};

export type LeaveStatus = 'pending' | 'approved' | 'rejected';
export type LeaveType = 'annual' | 'sick' | 'holiday';

export type LeaveRequest = {
  id: string;
  employee_id: string;
  employee_name: string;
  start_date: string;
  end_date: string;
	 leave_type: LeaveType;
  status: LeaveStatus;
  reviewed_by: string | null;
  reviewed_at: string | null;
  review_note: string | null;
  created_at: string;
  updated_at: string;
};

export type LeaveBalance = {
  id: string;
  employee_id: string;
	leave_year: number;
	leave_type: LeaveType;
  total_days: number;
  used_days: number;
  available_days: number;
};

export type AttendanceRecord = {
  id: string;
  employee_id: string;
  employee_name: string;
  department_id: string;
  department_name: string;
  attendance_date: string;
  check_in: string;
  check_out: string | null;
};

export type AttendanceReportRow = {
  employee_id: string;
  employee_name: string;
  department_id: string;
  department_name: string;
  present_days: number;
  completed_days: number;
  total_hours: number;
};
