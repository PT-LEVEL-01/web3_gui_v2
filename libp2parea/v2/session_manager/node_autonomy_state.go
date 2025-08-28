package session_manager

/*
节点自治状态管理
*/
type NodeAutonomyState struct {
	sessionManager     *SessionManager
	autonomyFinishChan chan bool
}

func NewNodeAutonomyState(sessionManager *SessionManager) *NodeAutonomyState {
	nas := NodeAutonomyState{
		sessionManager:     sessionManager,
		autonomyFinishChan: make(chan bool, 1),
	}
	return &nas
}

/*
设置自治完成
*/
func (this *NodeAutonomyState) SetAutonomyFinish() {
	if this.sessionManager.destroy.Load() {
		//销毁了则不放入了
		return
	}
	select {
	case this.autonomyFinishChan <- false:
	default:
	}
}

/*
等待网络自治完成，阻塞接口，需要等待
*/
func (this *NodeAutonomyState) WaitAutonomyFinish() {
	// 等待自治完成或area销毁信号
	select {
	case <-this.autonomyFinishChan: // 自治完成
		//取出来后立即放回去
		this.SetAutonomyFinish()
	case <-this.sessionManager.contextRoot.Done(): // area销毁,在p2p网络启动未成功时，如果调用销毁,避免卡住Destory逻辑
	}
}

/*
检查网络自治是否已经完成，立即返回，不等待
*/
func (this *NodeAutonomyState) CheckAutonomyFinish() (autoFinish bool) {
	// 等待自治完成或area销毁信号
	select {
	case <-this.autonomyFinishChan: // 自治完成
		//取出来后立即放回去
		this.SetAutonomyFinish()
		autoFinish = true
	case <-this.sessionManager.contextRoot.Done(): // area销毁,在p2p网络启动未成功时，如果调用销毁,避免卡住Destory逻辑
	default:
	}
	return autoFinish
}

/*
重置网络自治接口
*/
func (this *NodeAutonomyState) ResetAutonomyFinish() {
	select {
	case <-this.autonomyFinishChan:
	default:
	}
}

/*
设置状态为销毁
*/
func (this *NodeAutonomyState) Destroy() {
	//this.destroy.Store(true)
	select {
	case _, isopen := <-this.autonomyFinishChan: // 自治完成
		if isopen {
			close(this.autonomyFinishChan)
		}
	default:
		close(this.autonomyFinishChan)
	}
}
