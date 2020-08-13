lock '3.12.1'

set :application, 'limehd-syslog-server'
set :user, 'master'

set :repo_url, 'git@github.com:LimeHD/limehd-syslog-server.git' if ENV['USE_LOCAL_REPO'].nil?

set :linked_dirs, %w(log)

if ENV['BRANCH'].nil?
  ask :branch, proc { `git rev-parse --abbrev-ref HEAD`.chomp }
else
  set :branch, ENV['BRANCH']
end

set :deploy_to, -> { "/home/#{fetch(:user)}/#{fetch(:application)}" }

before 'systemd:daemon:validate', 'systemd:daemon:setup'
after 'deploy:publishing', 'systemd:daemon:restart'

namespace :deploy do
  after 'updated', :transfer_build
end

desc 'Transfer build'
task :transfer_build do
  on release_roles(:app) do
    upload! './bin', release_path, recursive: true
  end
end

before 'systemd:daemon:setup', :mkdir_user_systemd
task :mkdir_user_systemd do
  on release_roles(:app) do
    execute "mkdir -p /home/#{fetch(:user)}/.config/systemd/user"
  end
end
