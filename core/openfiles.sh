#Ubuntu server
echo -e "* hard nofile 655350 \n* soft nofile 655350" >> /etc/security/limits.conf
echo -e "session required pam_limits.so" >> /etc/pam.d/common-session && echo -e "session required pam_limits.so" >> /etc/pam.d/su
cp ./sysctl.conf /etc/sysctl.conf
sysctl -p
echo "set done, please reboot."
