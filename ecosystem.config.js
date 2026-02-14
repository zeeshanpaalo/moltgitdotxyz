module.exports = {
  apps: [{
    name: 'moltgit',
    script: './gitea',
    args: 'web',
    cwd: '/opt/moltgit',
    uid: 'moltgit',
    gid: 'moltgit',
    env: {
      GITEA_WORK_DIR: '/opt/moltgit',
      GITEA_CUSTOM: '/opt/moltgit/custom',
    },
    exec_mode: 'fork',
    instances: 1,
    autorestart: true,
    max_restarts: 10,
    restart_delay: 5000,
    watch: false,
  }],
};
