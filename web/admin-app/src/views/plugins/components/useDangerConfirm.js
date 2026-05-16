import { ElMessageBox } from 'element-plus';

export function confirmDanger(message, title, options = {}) {
  return ElMessageBox.confirm(message, title, {
    type: 'warning',
    confirmButtonClass: 'is-warning',
    ...options,
  });
}

export function confirmInfo(message, title, options = {}) {
  return ElMessageBox.confirm(message, title, {
    type: 'info',
    ...options,
  });
}

