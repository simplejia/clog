<?php

class CLog 
{
    private $ip;
    private $port;
    private $level;
    private $dbgcate;
    private $warcate;
    private $errcate;
    private $infocate;

    public function __construct($ip, $port, $level, $localip, $module, $subcate) {
        $this->dbgcate = implode(',', array($module, 'logdbg', $localip, $subcate));
        $this->warcate = implode(',', array($module, 'logwar', $localip, $subcate));
        $this->errcate = implode(',', array($module, 'logerr', $localip, $subcate));
        $this->infocate = implode(',', array($module, 'loginfo', $localip, $subcate));
        $this->ip = $ip;
        $this->port = $port;
        $this->level = $level;
    }   

    public function Log($cate, $content) {
        $sock = socket_create(AF_INET, SOCK_DGRAM, 0);
        if (!$sock) {
            $errorcode = socket_last_error();
            $errormsg = socket_strerror($errorcode);
            fprintf(STDERR, "CLog Log() [$errorcode] $errormsg\n");
            return;
        }   

        $out = $cate . ',' . $content;
        socket_sendto($sock, $out, strlen($out), 0, $this->ip, $this->port);
        socket_close($sock);
    }   

    public function Debug($content) {
        if (($this->level & 1) != 0) {
            $this->Log($this->dbgcate, $content);
        }
    }   

    public function Warn($content) {
        if (($this->level & 2) != 0) {
            $this->Log($this->warcate, $content);
        }
    }   

    public function Error($content) {
        if (($this->level & 4) != 0) {
            $this->Log($this->errcate, $content);
        }
    }   

    public function Info($content) {
        if (($this->level & 8) != 0) {
            $this->Log($this->infocate, $content);
        }
    }   
}

$module = "demo";
$ip = "127.0.0.1";
$port = 28702;
$level = 15;
$localip = '127.0.0.1'; // TODO replace with your localip

$clog = new CLog($ip, $port, $level, $localip, $module, "");
$clog->Debug("dbg msg");
$clog->Warn("war msg");
$clog->Error("err msg");
$clog->Info("info msg");

