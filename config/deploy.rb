lock '3.12.1'

set :application, 'epg_parsers'
set :user, 'master'

set :repo_url, 'git@github.com:LimeHD/epg_parsers.git' if ENV['USE_LOCAL_REPO'].nil?

set :linked_files, %w(.env)
set :linked_dirs, %w(log output)

if ENV['BRANCH'].nil?
  ask :branch, proc { `git rev-parse --abbrev-ref HEAD`.chomp }
else
  set :branch, ENV['BRANCH']
end

set :deploy_to, -> { "/home/#{fetch(:user)}/#{fetch(:application)}" }

namespace :deploy do
  after 'updated', :transfer_build
  after 'published', :reload_crontab
  after 'finishing_rollback', :reload_crontab
end

task :reload_crontab do
  on release_roles(:app) do
    execute "cd #{release_path}; crontab -u #{fetch(:user)} ./config/crontab"
  end
end

desc 'Transfer build'
task :transfer_build do
  on release_roles(:app) do
    upload! './bin', release_path, recursive: true
  end
end
