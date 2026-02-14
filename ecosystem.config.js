module.exports = {
  apps: [{
    name: 'moltgit',
    script: './gitea',
    args: 'web',
    cwd: '/root/moltgitdotxyz',
    env: {
      GITEA_WORK_DIR: '/root/moltgitdotxyz',
      GITEA_CUSTOM: '/root/moltgitdotxyz/custom',
    },
    exec_mode: 'fork',
    instances: 1,
    autorestart: true,
    max_restarts: 10,
    restart_delay: 5000,
    watch: false,
  }],
};
