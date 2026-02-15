module.exports = {
  apps: [{
    name: 'moltgit',
    script: './moltgit',
    args: 'web',
    cwd: '/opt/moltgit',
    env: {
      GITEA_WORK_DIR: '/opt/moltgit',
      GITEA_CUSTOM: '/opt/moltgit/custom',
      GITEA_I_AM_BEING_UNSAFE_RUNNING_AS_ROOT: 'true',
    },
    exec_mode: 'fork',
    instances: 1,
    autorestart: true,
    max_restarts: 10,
    restart_delay: 5000,
    watch: false,
  }],
};
